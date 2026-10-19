package controller

import (
	"encoding/json"
	"fmt"
	"testing"

	"gofiber-workshop/models"
)

func TestDogTypeBoundaries(t *testing.T) {
	cases := []struct {
		id   int
		want string
	}{
		{-1, "no color"}, {0, "no color"}, {9, "no color"},
		{10, "red"}, {35, "red"}, {50, "red"},
		{51, "no color"}, {99, "no color"},
		{100, "green"}, {125, "green"}, {150, "green"}, {151, "no color"},
		{199, "no color"}, {200, "pink"}, {210, "pink"}, {250, "pink"},
		{251, "no color"}, {999, "no color"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.id), func(t *testing.T) {
			if got := dogType(tc.id); got != tc.want {
				t.Fatalf("dogType(%d) = %q, want %q", tc.id, got, tc.want)
			}
		})
	}
}

func TestDogSummaryWorkshopFixtures(t *testing.T) {
	ids := []int{9, 10, 50, 51, 99, 100, 150, 151, 200, 250, 251, 999}
	dogs := make([]models.Dogs, 0, len(ids))
	for i, id := range ids {
		dogs = append(dogs, models.Dogs{Name: fmt.Sprintf("Dog%02d", i+1), DogID: id})
	}
	result := buildDogSummary(dogs)
	if result.Name != "golang-test" || result.Count != 12 || len(result.Data) != 12 {
		t.Fatalf("unexpected response: %+v", result)
	}
	if result.SumRed != 2 || result.SumGreen != 2 || result.SumPink != 2 || result.SumNoColor != 6 {
		t.Fatalf("incorrect color counts: %+v", result)
	}
	if result.SumRed+result.SumGreen+result.SumPink+result.SumNoColor != result.Count {
		t.Fatal("color totals must equal count")
	}
	for i, dog := range result.Data {
		if dog.Name != dogs[i].Name || dog.DogID != dogs[i].DogID || dog.Type != dogType(dogs[i].DogID) {
			t.Fatalf("incorrect dog response at index %d: %+v", i, dog)
		}
	}
}

func TestDogSummaryCountsRowsNotDogIDs(t *testing.T) {
	result := buildDogSummary([]models.Dogs{{DogID: 10}, {DogID: 10}, {DogID: 200}})
	if result.SumRed != 2 || result.SumPink != 1 || result.Count != 3 {
		t.Fatalf("must count each row once, including duplicate dog_id: %+v", result)
	}
}

func TestDogSummaryEmptyArray(t *testing.T) {
	for _, dogs := range [][]models.Dogs{nil, {}} {
		result := buildDogSummary(dogs)
		body, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]interface{}
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded["data"] == nil || result.Count != 0 || result.SumRed+result.SumGreen+result.SumPink+result.SumNoColor != 0 {
			t.Fatalf("empty data must be [] with zero counts: %s", body)
		}
	}
}

func TestDogRequestValidation(t *testing.T) {
	for _, tc := range []struct {
		body string
		ok   bool
	}{
		{`{"name":"Dog","dog_id":0}`, true},
		{`{"name":"Dog","dog_id":10}`, true},
		{`{"name":"Dog"}`, false},
		{`{"name":"Dog","dog_id":null}`, false},
		{`{"name":"Dog","dog_id":-1}`, false},
		{`{"name":"","dog_id":10}`, false},
	} {
		var request models.DogRequest
		if err := json.Unmarshal([]byte(tc.body), &request); err != nil {
			t.Fatal(err)
		}
		if ok := validate.Struct(request) == nil; ok != tc.ok {
			t.Errorf("validation %s: got %v, want %v", tc.body, ok, tc.ok)
		}
	}
}

func TestParseID(t *testing.T) {
	for _, invalid := range []string{"", "0", "-1", "abc", "1.5", "1 OR 1=1", "18446744073709551616"} {
		if _, err := parseID(invalid); err == nil {
			t.Errorf("accepted invalid id %q", invalid)
		}
	}
	if id, err := parseID("12"); err != nil || id != 12 {
		t.Fatalf("valid id rejected: %d, %v", id, err)
	}
}
