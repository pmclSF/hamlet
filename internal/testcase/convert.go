package testcase

import "github.com/pmclSF/hamlet/internal/models"

// ToModel converts a testcase.TestCase to models.TestCase for snapshot storage.
func ToModel(tc *TestCase) models.TestCase {
	var param *models.ParameterizationInfo
	if tc.Parameterized != nil {
		param = &models.ParameterizationInfo{
			IsTemplate:         tc.Parameterized.IsTemplate,
			ParamSignature:     tc.Parameterized.ParamSignature,
			EstimatedInstances: tc.Parameterized.EstimatedInstances,
		}
	}

	return models.TestCase{
		TestID:            tc.TestID,
		CanonicalIdentity: tc.CanonicalIdentity,
		FilePath:          tc.FilePath,
		SuiteHierarchy:    tc.SuiteHierarchy,
		TestName:          tc.TestName,
		Framework:         tc.Framework,
		Language:          tc.Language,
		Line:              tc.Line,
		ExtractionKind:    string(tc.ExtractionKind),
		Confidence:        tc.Confidence,
		Parameterized:     param,
	}
}

// ToModels converts a slice of testcase.TestCase to models.TestCase.
func ToModels(cases []TestCase) []models.TestCase {
	if len(cases) == 0 {
		return nil
	}
	result := make([]models.TestCase, len(cases))
	for i := range cases {
		result[i] = ToModel(&cases[i])
	}
	return result
}
