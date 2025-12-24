package handler

import (
	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"
	fileanalysis "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/clients/fileanalysis"
)

func reportIDFromSubmissionID(submissionID string) string {
	if submissionID == "" {
		return ""
	}
	return "rep-" + submissionID
}

func mapAnalysisReport(report *fileanalysis.ReportResponse) *api.PlagiarismReport {
	if report == nil {
		return nil
	}
	return &api.PlagiarismReport{
		ReportId:           reportIDFromSubmissionID(report.SubmissionId),
		Status:             api.PlagiarismReportStatus(report.Status),
		PlagiarismDetected: report.PlagiarismDetected,
		SimilarityPercent:  report.SimilarityPercent,
		CreatedAt:          report.CreatedAt,
		CompletedAt:        report.CompletedAt,
		ErrorMessage:       report.ErrorMessage,
		MatchedSubmissions: mapMatchedSubmissions(report.MatchedSubmissions),
	}
}

func mapMatchedSubmissions(matches *[]fileanalysis.MatchedSubmission) *[]api.MatchedSubmission {
	if matches == nil {
		return nil
	}
	result := make([]api.MatchedSubmission, 0, len(*matches))
	for _, match := range *matches {
		result = append(result, api.MatchedSubmission{
			SubmissionId:      match.SubmissionId,
			MatchedChunks:     match.MatchedChunks,
			SimilarityPercent: float32(match.SimilarityPercent),
		})
	}
	return &result
}

func mapWorkReportItem(report fileanalysis.WorkReportItem) api.WorkReportItem {
	return api.WorkReportItem{
		ReportId:           reportIDFromSubmissionID(report.SubmissionId),
		SubmissionId:       report.SubmissionId,
		Status:             api.WorkReportItemStatus(report.Status),
		PlagiarismDetected: report.PlagiarismDetected,
		SimilarityPercent:  report.SimilarityPercent,
		CreatedAt:          report.CreatedAt,
		CompletedAt:        report.CompletedAt,
	}
}
