package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type NewsItem struct {
	Date                           string            `json:"date"`
	PolicyName                     string            `json:"policy_name"`
	Department                     string            `json:"department"`
	KeyPoints                      []string          `json:"key_points,omitempty"`
	BudgetAllocation               float64           `json:"budget_allocation,omitempty"`
	PublicApprovalRating           float64           `json:"public_approval_rating,omitempty"`
	ImplementationPhase            string            `json:"implementation_phase,omitempty"`
	AffectedIndustries             []string          `json:"affected_industries,omitempty"`
	EstimatedJobCreation           int               `json:"estimated_job_creation,omitempty"`
	ProjectedCarbonReduction       string            `json:"projected_carbon_reduction,omitempty"`
	Challenges                     []string          `json:"challenges,omitempty"`
	NextReviewDate                 string            `json:"next_review_date,omitempty"`
	DailySolarInstallations        int               `json:"daily_solar_installations,omitempty"`
	DailyEVPurchases               int               `json:"daily_ev_purchases,omitempty"`
	EnergyEfficiencyComplianceRate float64           `json:"energy_efficiency_compliance_rate,omitempty"`
	PublicInquiriesReceived        int               `json:"public_inquiries_received,omitempty"`
	MediaMentions                  int               `json:"media_mentions,omitempty"`
	StockMarketImpact              map[string]string `json:"stock_market_impact,omitempty"`
	LocalGovernmentAdoptionRate    string            `json:"local_government_adoption_rate,omitempty"`
	NewlyIdentifiedChallenges      []string          `json:"newly_identified_challenges,omitempty"`
}

func main() {
	http.HandleFunc("/summarize", handleSummarize)
	http.HandleFunc("/", handleIndex)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/index.html"))
	tmpl.Execute(w, nil)
}

func handleSummarize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := readJSON("sample.json")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading JSON: %v", err), http.StatusInternalServerError)
		return
	}

	summary, err := summarizeWithGemini(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error summarizing data: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"original_data": json.RawMessage(data),
		"summary":       summary,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func readJSON(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

func summarizeWithGemini(data []byte) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("AIzaSyBQ-bfTfxdGJYTvFtodFUiKLPdJekCyi6M"))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro")

	var newsItems []NewsItem
	err = json.Unmarshal(data, &newsItems)
	if err != nil {
		return "", err
	}

	// Create a summarized version of the news data to send as a prompt
	var summarizedData string
	for _, item := range newsItems {
		summarizedData += fmt.Sprintf("Policy: %s\nDate: %s\nDepartment: %s\nKey Points: %v\nBudget: $%v\nApproval: %v%%\nPhase: %s\nIndustries: %v\nJobs: %d\nCarbon Reduction: %s\nChallenges: %v\nReview Date: %s\n\n",
			item.PolicyName, item.Date, item.Department, item.KeyPoints, item.BudgetAllocation, item.PublicApprovalRating, item.ImplementationPhase, item.AffectedIndustries, item.EstimatedJobCreation, item.ProjectedCarbonReduction, item.Challenges, item.NextReviewDate)
		if item.DailySolarInstallations != 0 || item.DailyEVPurchases != 0 || item.EnergyEfficiencyComplianceRate != 0 || item.PublicInquiriesReceived != 0 {
			summarizedData += fmt.Sprintf("Daily Solar Installations: %d\nDaily EV Purchases: %d\nEfficiency Rate: %v%%\nInquiries: %d\nMentions: %d\nStock Impact: %v\nAdoption Rate: %s\nNew Challenges: %v\n\n",
				item.DailySolarInstallations, item.DailyEVPurchases, item.EnergyEfficiencyComplianceRate, item.PublicInquiriesReceived, item.MediaMentions, item.StockMarketImpact, item.LocalGovernmentAdoptionRate, item.NewlyIdentifiedChallenges)
		}
	}

	prompt := fmt.Sprintf("Analyze the provided JSON data and create a meaningful summary. The data represents information about government policies, bills, and expenditures. Here's the data:\n%s", summarizedData)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	var summary string
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				summary += fmt.Sprintf("%v", part)
			}
		}
	}

	return summary, nil
}
