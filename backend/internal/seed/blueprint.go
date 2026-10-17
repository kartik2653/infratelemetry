package seed

import "chargingdatainfra/backend/internal/domain"

type connSpec struct {
	Type     domain.ChargerType
	Standard string
	KW       float64
}

type stationSpec struct {
	ID, Name, Area, Address string
	Lat, Lng                float64
	Conns                   []connSpec
	MeanDC, MeanAC          float64
	WeekdayMul, WeekendMul  float64
	Weight                  string
	PayFail                 float64
	Interrupt               float64
	HogAC                   float64
	HogDC                   float64
	DerateFromDay           int
	PeakTariff              float64
	OffPeakTariff           float64
	PeakCost                float64
	OffPeakCost             float64
}

func blueprint() []stationSpec {
	return []stationSpec{
		{
			ID: "BLR-001", Name: "Indiranagar Metro Deck", Area: "Indiranagar",
			Address: "100 Feet Road, Indiranagar, Bengaluru 560038", Lat: 12.9784, Lng: 77.6408,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 60}, {domain.ChargerDC, "CCS2", 60}, {domain.ChargerAC, "Type2", 22}},
			MeanDC: 13, MeanAC: 4.5, WeekdayMul: 1.05, WeekendMul: 0.85, Weight: "commute",
			PayFail: 0.012, Interrupt: 0.012,
			PeakTariff: 22, OffPeakTariff: 18, PeakCost: 11, OffPeakCost: 8.2,
		},
		{
			ID: "BLR-002", Name: "Koramangala 5th Block Hub", Area: "Koramangala",
			Address: "80 Feet Road, Koramangala 5th Block, Bengaluru 560095", Lat: 12.9352, Lng: 77.6245,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 50}, {domain.ChargerDC, "CCS2", 50}, {domain.ChargerAC, "Type2", 7.4}, {domain.ChargerAC, "Type2", 7.4}},
			MeanDC: 6.5, MeanAC: 3.4, WeekdayMul: 0.95, WeekendMul: 1.05, Weight: "evening",
			PayFail: 0.012, Interrupt: 0.012,
			PeakTariff: 24, OffPeakTariff: 18, PeakCost: 10.5, OffPeakCost: 8,
		},
		{
			ID: "BLR-003", Name: "Whitefield ITPL Gate", Area: "Whitefield",
			Address: "ITPL Main Road, Whitefield, Bengaluru 560066", Lat: 12.9698, Lng: 77.7500,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 120}, {domain.ChargerDC, "CCS2", 120}},
			MeanDC: 12, MeanAC: 0, WeekdayMul: 1.25, WeekendMul: 0.28, Weight: "office",
			PayFail: 0.01, Interrupt: 0.01,
			PeakTariff: 23, OffPeakTariff: 18, PeakCost: 11.5, OffPeakCost: 8.4,
		},
		{
			ID: "BLR-004", Name: "Electronic City Phase 1", Area: "Electronic City",
			Address: "Hosur Road, Electronics City Phase 1, Bengaluru 560100", Lat: 12.8452, Lng: 77.6602,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 60}, {domain.ChargerDC, "CCS2", 60}, {domain.ChargerAC, "Type2", 22}},
			MeanDC: 1.15, MeanAC: 0.7, WeekdayMul: 1, WeekendMul: 0.8, Weight: "flat",
			PayFail: 0.015, Interrupt: 0.01,
			PeakTariff: 21, OffPeakTariff: 17, PeakCost: 10.5, OffPeakCost: 8,
		},
		{
			ID: "BLR-005", Name: "Jayanagar 4th Block", Area: "Jayanagar",
			Address: "11th Main, Jayanagar 4th Block, Bengaluru 560011", Lat: 12.9308, Lng: 77.5838,
			Conns:  []connSpec{{domain.ChargerAC, "Type2", 7.4}, {domain.ChargerAC, "Type2", 7.4}, {domain.ChargerAC, "Type2", 7.4}},
			MeanDC: 0, MeanAC: 3.1, WeekdayMul: 0.9, WeekendMul: 1.15, Weight: "evening",
			PayFail: 0.01, Interrupt: 0.008, HogAC: 0.05,
			PeakTariff: 18, OffPeakTariff: 15, PeakCost: 9.2, OffPeakCost: 7.4,
		},
		{
			ID: "BLR-006", Name: "Hebbal Flyover Lot", Area: "Hebbal",
			Address: "Bellary Road, Hebbal, Bengaluru 560024", Lat: 13.0358, Lng: 77.5970,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 50}, {domain.ChargerDC, "CCS2", 50}, {domain.ChargerAC, "Type2", 22}},
			MeanDC: 5.5, MeanAC: 2.2, WeekdayMul: 1, WeekendMul: 0.75, Weight: "commute",
			PayFail: 0.02, Interrupt: 0.07,
			PeakTariff: 22, OffPeakTariff: 18, PeakCost: 11, OffPeakCost: 8.4,
		},
		{
			ID: "BLR-007", Name: "Yeshwanthpur Depot", Area: "Yeshwanthpur",
			Address: "Tumkur Road, Yeshwanthpur, Bengaluru 560022", Lat: 13.0280, Lng: 77.5409,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 150}, {domain.ChargerDC, "CCS2", 150}},
			MeanDC: 14, MeanAC: 0, WeekdayMul: 1.1, WeekendMul: 0.7, Weight: "busy",
			PayFail: 0.012, Interrupt: 0.012,
			PeakTariff: 21, OffPeakTariff: 17, PeakCost: 19.5, OffPeakCost: 15.5,
		},
		{
			ID: "BLR-008", Name: "MG Road Parade", Area: "MG Road",
			Address: "Mahatma Gandhi Road, Bengaluru 560001", Lat: 12.9756, Lng: 77.6068,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 60}, {domain.ChargerAC, "Type2", 22}, {domain.ChargerAC, "Type2", 22}},
			MeanDC: 4.2, MeanAC: 2.4, WeekdayMul: 1.05, WeekendMul: 0.95, Weight: "commute",
			PayFail: 0.01, Interrupt: 0.008,
			PeakTariff: 34, OffPeakTariff: 26, PeakCost: 10, OffPeakCost: 8,
		},
		{
			ID: "BLR-009", Name: "Bannerghatta Road", Area: "Bannerghatta",
			Address: "Bannerghatta Road, Bengaluru 560076", Lat: 12.8876, Lng: 77.5970,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 50}, {domain.ChargerDC, "CHAdeMO", 50}},
			MeanDC: 2.4, MeanAC: 0, WeekdayMul: 0.4, WeekendMul: 3.8, Weight: "weekend_bias",
			PayFail: 0.012, Interrupt: 0.01,
			PeakTariff: 22, OffPeakTariff: 17, PeakCost: 10.5, OffPeakCost: 8,
		},
		{
			ID: "BLR-010", Name: "Marathahalli Bridge", Area: "Marathahalli",
			Address: "Outer Ring Road, Marathahalli, Bengaluru 560037", Lat: 12.9591, Lng: 77.6974,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 60}, {domain.ChargerDC, "CCS2", 60}, {domain.ChargerAC, "Type2", 11}},
			MeanDC: 6.2, MeanAC: 2.4, WeekdayMul: 1.1, WeekendMul: 0.65, Weight: "commute",
			PayFail: 0.18, Interrupt: 0.015,
			PeakTariff: 23, OffPeakTariff: 18, PeakCost: 11, OffPeakCost: 8.3,
		},
		{
			ID: "BLR-011", Name: "HSR Layout 27th Main", Area: "HSR Layout",
			Address: "27th Main, HSR Layout, Bengaluru 560102", Lat: 12.9121, Lng: 77.6446,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 50}, {domain.ChargerDC, "CCS2", 50}, {domain.ChargerAC, "Type2", 22}},
			MeanDC: 10, MeanAC: 4, WeekdayMul: 1, WeekendMul: 0.9, Weight: "commute",
			PayFail: 0.012, Interrupt: 0.01,
			PeakTariff: 23, OffPeakTariff: 18, PeakCost: 10.5, OffPeakCost: 8,
		},
		{
			ID: "BLR-012", Name: "Malleshwaram 8th Cross", Area: "Malleshwaram",
			Address: "8th Cross, Malleshwaram, Bengaluru 560003", Lat: 13.0035, Lng: 77.5647,
			Conns:  []connSpec{{domain.ChargerAC, "Type2", 11}, {domain.ChargerAC, "Type2", 11}, {domain.ChargerAC, "Type2", 11}, {domain.ChargerAC, "Type2", 11}},
			MeanDC: 0, MeanAC: 2.6, WeekdayMul: 0.95, WeekendMul: 1.1, Weight: "evening",
			PayFail: 0.01, Interrupt: 0.008, HogAC: 0.22,
			PeakTariff: 17, OffPeakTariff: 14.5, PeakCost: 9, OffPeakCost: 7.2,
		},
		{
			ID: "BLR-013", Name: "Bellandur Outer Ring", Area: "Bellandur",
			Address: "Outer Ring Road, Bellandur, Bengaluru 560103", Lat: 12.9260, Lng: 77.6762,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 120}, {domain.ChargerDC, "CCS2", 120}, {domain.ChargerAC, "Type2", 22}},
			MeanDC: 11, MeanAC: 3.2, WeekdayMul: 1.2, WeekendMul: 0.35, Weight: "office",
			PayFail: 0.012, Interrupt: 0.012,
			PeakTariff: 24, OffPeakTariff: 18, PeakCost: 12, OffPeakCost: 9,
		},
		{
			ID: "BLR-014", Name: "Rajajinagar Chord Road", Area: "Rajajinagar",
			Address: "Chord Road, Rajajinagar, Bengaluru 560010", Lat: 12.9915, Lng: 77.5520,
			Conns:  []connSpec{{domain.ChargerAC, "Type2", 22}, {domain.ChargerAC, "Type2", 22}, {domain.ChargerDC, "CCS2", 50}},
			MeanDC: 1.6, MeanAC: 1.3, WeekdayMul: 1, WeekendMul: 0.85, Weight: "flat",
			PayFail: 0.015, Interrupt: 0.012,
			PeakTariff: 20, OffPeakTariff: 16, PeakCost: 10, OffPeakCost: 7.8,
		},
		{
			ID: "BLR-015", Name: "JP Nagar 24th Main", Area: "JP Nagar",
			Address: "24th Main, JP Nagar 6th Phase, Bengaluru 560078", Lat: 12.9063, Lng: 77.5857,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 50}, {domain.ChargerDC, "CCS2", 50}, {domain.ChargerAC, "Type2", 7.4}},
			MeanDC: 4.4, MeanAC: 2, WeekdayMul: 1, WeekendMul: 0.9, Weight: "commute",
			PayFail: 0.02, Interrupt: 0.03,
			PeakTariff: 21, OffPeakTariff: 17, PeakCost: 10.5, OffPeakCost: 8,
		},
		{
			ID: "BLR-016", Name: "Sarjapur Road Junction", Area: "Sarjapur",
			Address: "Sarjapur Road, Bengaluru 560035", Lat: 12.9102, Lng: 77.6870,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 60}, {domain.ChargerAC, "Type2", 11}, {domain.ChargerAC, "Type2", 11}},
			MeanDC: 4.6, MeanAC: 2.2, WeekdayMul: 1.05, WeekendMul: 0.8, Weight: "commute",
			PayFail: 0.012, Interrupt: 0.012, DerateFromDay: 78,
			PeakTariff: 22, OffPeakTariff: 17.5, PeakCost: 10.5, OffPeakCost: 8.1,
		},
		{
			ID: "BLR-017", Name: "Yelahanka New Town", Area: "Yelahanka",
			Address: "New Town Main Road, Yelahanka, Bengaluru 560064", Lat: 13.1007, Lng: 77.5963,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 50}, {domain.ChargerDC, "CCS2", 50}},
			MeanDC: 0.95, MeanAC: 0, WeekdayMul: 1, WeekendMul: 0.7, Weight: "flat",
			PayFail: 0.012, Interrupt: 0.008,
			PeakTariff: 20, OffPeakTariff: 16, PeakCost: 10, OffPeakCost: 7.8,
		},
		{
			ID: "BLR-018", Name: "KR Puram Highway", Area: "KR Puram",
			Address: "Old Madras Road, KR Puram, Bengaluru 560036", Lat: 13.0005, Lng: 77.6947,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 150}, {domain.ChargerDC, "CHAdeMO", 50}},
			MeanDC: 3.4, MeanAC: 0, WeekdayMul: 1, WeekendMul: 1.15, Weight: "highway",
			PayFail: 0.015, Interrupt: 0.015, HogDC: 0.06,
			PeakTariff: 26, OffPeakTariff: 20, PeakCost: 12, OffPeakCost: 9,
		},
		{
			ID: "BLR-019", Name: "Basavanagudi South End", Area: "Basavanagudi",
			Address: "Gandhi Bazaar, Basavanagudi, Bengaluru 560004", Lat: 12.9422, Lng: 77.5754,
			Conns:  []connSpec{{domain.ChargerAC, "Type2", 22}, {domain.ChargerAC, "Type2", 22}, {domain.ChargerAC, "Type2", 22}},
			MeanDC: 0, MeanAC: 3.3, WeekdayMul: 1, WeekendMul: 1.05, Weight: "evening",
			PayFail: 0.008, Interrupt: 0.006,
			PeakTariff: 18, OffPeakTariff: 15, PeakCost: 9.2, OffPeakCost: 7.4,
		},
		{
			ID: "BLR-020", Name: "Manyata Tech Park", Area: "Manyata",
			Address: "Nagavara, Manyata Tech Park, Bengaluru 560045", Lat: 13.0485, Lng: 77.6210,
			Conns:  []connSpec{{domain.ChargerDC, "CCS2", 60}, {domain.ChargerDC, "CCS2", 60}, {domain.ChargerDC, "CCS2", 60}, {domain.ChargerAC, "Type2", 22}},
			MeanDC: 13, MeanAC: 4, WeekdayMul: 1.3, WeekendMul: 0.22, Weight: "office",
			PayFail: 0.01, Interrupt: 0.01,
			PeakTariff: 25, OffPeakTariff: 19, PeakCost: 12, OffPeakCost: 8.8,
		},
	}
}
