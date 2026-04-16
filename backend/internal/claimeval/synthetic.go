package claimeval

import "fmt"

// SyntheticScenario returns realistic sample inputs for the evaluation layer.
func SyntheticScenario(name string) (SyntheticSource, error) {
	switch name {
	case "genuine_worker":
		return genuineWorkerScenario(), nil
	case "lazy_fraud":
		return lazyFraudScenario(), nil
	case "smart_fraud":
		return smartFraudScenario(), nil
	default:
		return SyntheticSource{}, fmt.Errorf("unknown synthetic scenario %q", name)
	}
}

func genuineWorkerScenario() SyntheticSource {
	active := true
	login := 4.5
	attempted := 7
	completed := 1
	actual := 180.0
	expected := 540.0
	return SyntheticSource{
		WorkerID:         101,
		Zone:             "Tambaram, Chennai",
		SegmentID:        "tambaram-chennai:b:two-wheeler",
		ActiveBefore:     &active,
		ActiveDuring:     &active,
		LoginDuration:    &login,
		OrdersAttempted:  &attempted,
		OrdersCompleted:  &completed,
		EarningsActual:   &actual,
		EarningsExpected: &expected,
	}
}

func lazyFraudScenario() SyntheticSource {
	inactive := false
	login := 0.0
	attempted := 0
	completed := 0
	actual := 0.0
	expected := 420.0
	return SyntheticSource{
		WorkerID:         202,
		Zone:             "Rohini, Delhi",
		SegmentID:        "rohini-delhi:b:two-wheeler",
		ActiveBefore:     &inactive,
		ActiveDuring:     &inactive,
		LoginDuration:    &login,
		OrdersAttempted:  &attempted,
		OrdersCompleted:  &completed,
		EarningsActual:   &actual,
		EarningsExpected: &expected,
	}
}

func smartFraudScenario() SyntheticSource {
	active := true
	login := 5.0
	attempted := 6
	completed := 4
	actual := 40.0
	expected := 610.0
	return SyntheticSource{
		WorkerID:         303,
		Zone:             "Adyar, Chennai",
		SegmentID:        "adyar-chennai:b:two-wheeler",
		ActiveBefore:     &active,
		ActiveDuring:     &active,
		LoginDuration:    &login,
		OrdersAttempted:  &attempted,
		OrdersCompleted:  &completed,
		EarningsActual:   &actual,
		EarningsExpected: &expected,
	}
}
