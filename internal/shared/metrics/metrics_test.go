package metrics

import (
	"math"
	"strings"
	"testing"
)

const epsilon = 1e-12

func assertApplicable(t *testing.T, r MetricResult, want float64) {
	t.Helper()

	if !r.Applicable {
		t.Fatalf("%s: not applicable (%s), want value %v", r.Name, r.Reason, want)
	}

	if math.Abs(r.Value-want) > epsilon {
		t.Fatalf("%s: value %v, want %v", r.Name, r.Value, want)
	}
}

func assertNotApplicable(t *testing.T, r MetricResult) {
	t.Helper()

	if r.Applicable {
		t.Fatalf("%s: applicable with value %v, want not applicable", r.Name, r.Value)
	}

	if r.Reason == "" {
		t.Fatalf("%s: not applicable without a reason", r.Name)
	}

	if r.Definition == "" {
		t.Fatalf("%s: missing definition", r.Name)
	}
}

func TestAMC(t *testing.T) {
	assertNotApplicable(t, AMC(0, 0))
	assertApplicable(t, AMC(5, 2), 2.5)
	assertApplicable(t, AMC(1, 1), 1)
}

func TestLCOM1(t *testing.T) {
	assertNotApplicable(t, LCOM1(0, 0, 1, 3)) // fewer than 2 methods
	assertNotApplicable(t, LCOM1(1, 0, 2, 0)) // no fields
	assertApplicable(t, LCOM1(3, 0, 3, 2), 3)
	assertApplicable(t, LCOM1(1, 2, 3, 2), 0) // clamped at zero
}

func TestLCOM96b(t *testing.T) {
	assertNotApplicable(t, LCOM96b(0, 0, 3))
	assertNotApplicable(t, LCOM96b(0, 3, 0))
	assertApplicable(t, LCOM96b(3, 3, 3), 1-1.0/3)
	assertApplicable(t, LCOM96b(9, 3, 3), 0) // full matrix
	assertApplicable(t, LCOM96b(1, 1, 1), 0) // defined at a single method
	assertApplicable(t, LCOM96b(0, 2, 4), 1) // empty matrix
}

func TestTCC(t *testing.T) {
	assertNotApplicable(t, TCC(0, 0))
	assertNotApplicable(t, TCC(0, 1))
	assertApplicable(t, TCC(1, 3), 1.0/3)
	assertApplicable(t, TCC(3, 3), 1)
	assertApplicable(t, TCC(0, 2), 0)
}

func TestCAMC(t *testing.T) {
	assertNotApplicable(t, CAMC(0, 0, 2))
	assertNotApplicable(t, CAMC(0, 2, 0))
	assertApplicable(t, CAMC(2, 3, 2), 2.0/6)
	assertApplicable(t, CAMC(4, 2, 2), 1)
}

func TestCBO(t *testing.T) {
	assertApplicable(t, CBO(0), 0) // always applicable, even at zero
	assertApplicable(t, CBO(7), 7)
}

func TestAbstractness(t *testing.T) {
	assertNotApplicable(t, Abstractness(0, 0))
	assertApplicable(t, Abstractness(1, 4), 0.25)
	assertApplicable(t, Abstractness(0, 3), 0)
}

func TestInstability(t *testing.T) {
	// Isolated package: defined as maximally stable, with a reason.
	isolated := Instability(0, 0)
	assertApplicable(t, isolated, 0)

	if isolated.Reason == "" {
		t.Fatal("isolated instability should carry the defined-as-0 reason")
	}

	assertApplicable(t, Instability(1, 0), 0)
	assertApplicable(t, Instability(0, 2), 1)
	assertApplicable(t, Instability(1, 3), 0.75)
}

func TestDistance(t *testing.T) {
	abstractness := Abstractness(1, 4) // 0.25
	instability := Instability(1, 3)   // 0.75
	assertApplicable(t, Distance(abstractness, instability), 0)

	assertNotApplicable(t, Distance(Abstractness(0, 0), instability))
	// Isolated packages have instability 0, so distance stays computable.
	assertApplicable(t, Distance(abstractness, Instability(0, 0)), 0.75)
	assertApplicable(t, Distance(Abstractness(1, 1), Instability(0, 2)), 1)
}

func TestReusabilityAllComponents(t *testing.T) {
	weights := DefaultReusabilityWeights()
	r := Reusability(
		ReusabilityComponent{Name: ComponentCohesion, Value: 1, Applicable: true},
		ReusabilityComponent{Name: ComponentCoupling, Value: 1, Applicable: true},
		ReusabilityComponent{Name: ComponentTestability, Value: 1, Applicable: true},
		ReusabilityComponent{Name: ComponentDocumentation, Value: 1, Applicable: true},
		weights,
	)
	assertApplicable(t, r, 1)

	if r.Reason != "" {
		t.Fatalf("no dropped components expected, got reason %q", r.Reason)
	}
}

func TestReusabilityRenormalization(t *testing.T) {
	weights := DefaultReusabilityWeights()
	// Cohesion dropped: remaining weights 0.25+0.25+0.15 renormalize to 1.
	r := Reusability(
		ReusabilityComponent{Name: ComponentCohesion},
		ReusabilityComponent{Name: ComponentCoupling, Value: 0.5, Applicable: true},
		ReusabilityComponent{Name: ComponentTestability, Value: 1, Applicable: true},
		ReusabilityComponent{Name: ComponentDocumentation, Value: 0, Applicable: true},
		weights,
	)
	want := (0.25*0.5 + 0.25*1 + 0.15*0) / 0.65
	assertApplicable(t, r, want)

	if !strings.Contains(r.Reason, ComponentCohesion) {
		t.Fatalf("reason %q does not list the dropped component", r.Reason)
	}
}

func TestReusabilityAllDropped(t *testing.T) {
	r := Reusability(
		ReusabilityComponent{Name: ComponentCohesion, Reason: "type has no methods"},
		ReusabilityComponent{Name: ComponentCoupling, Reason: "no dependency data"},
		ReusabilityComponent{Name: ComponentTestability, Reason: "type has no methods"},
		ReusabilityComponent{Name: ComponentDocumentation, Reason: "type has no exported members"},
		DefaultReusabilityWeights(),
	)
	assertNotApplicable(t, r)
	// The reason names every dropped component with its own cause.
	for _, want := range []string{
		"every component dropped",
		"cohesion (type has no methods)",
		"coupling (no dependency data)",
		"testability (type has no methods)",
		"documentation (type has no exported members)",
	} {
		if !strings.Contains(r.Reason, want) {
			t.Fatalf("reason %q missing %q", r.Reason, want)
		}
	}
}

func TestReusabilityComponents(t *testing.T) {
	if c := CohesionComponent(LCOM96b(3, 3, 3)); !c.Applicable || math.Abs(c.Value-1.0/3) > epsilon {
		t.Fatalf("cohesion component = %+v", c)
	}

	if c := CohesionComponent(LCOM96b(0, 0, 3)); c.Applicable {
		t.Fatalf("cohesion component should drop when LCOM96b is n/a")
	}

	if c := CouplingComponent(0); !c.Applicable || c.Value != 1 {
		t.Fatalf("coupling component at CBO=0 = %+v", c)
	}

	if c := CouplingComponent(3); math.Abs(c.Value-0.25) > epsilon {
		t.Fatalf("coupling component at CBO=3 = %+v", c)
	}

	if c := TestabilityComponent(AMC(1, 1)); !c.Applicable || c.Value != 1 {
		t.Fatalf("testability component at AMC=1 = %+v", c)
	}

	if c := TestabilityComponent(AMC(3, 1)); math.Abs(c.Value-1.0/3) > epsilon {
		t.Fatalf("testability component at AMC=3 = %+v", c)
	}

	if c := TestabilityComponent(AMC(0, 0)); c.Applicable {
		t.Fatalf("testability component should drop when AMC is n/a")
	}

	if c := DocumentationComponent(2, 4); !c.Applicable || c.Value != 0.5 {
		t.Fatalf("documentation component = %+v", c)
	}

	if c := DocumentationComponent(0, 0); c.Applicable {
		t.Fatalf("documentation component should drop with no exported members")
	}
}

func TestWeightsValidate(t *testing.T) {
	err := DefaultReusabilityWeights().Validate()
	if err != nil {
		t.Fatal(err)
	}

	err = (ReusabilityWeights{Cohesion: -1}).Validate()
	if err == nil {
		t.Fatal("negative weight accepted")
	}

	err = (ReusabilityWeights{}).Validate()
	if err == nil {
		t.Fatal("all-zero weights accepted")
	}
}

func TestClosure(t *testing.T) {
	got := Closure([]string{MetricReusability})

	want := []string{MetricAMC, MetricLCOM96b, MetricCBO, MetricReusability}
	if len(got) != len(want) {
		t.Fatalf("closure = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("closure = %v, want %v", got, want)
		}
	}

	got = Closure([]string{MetricDistance})

	want = []string{MetricAbstractness, MetricInstability, MetricDistance}
	if len(got) != len(want) {
		t.Fatalf("closure = %v, want %v", got, want)
	}

	if got := Closure([]string{MetricTCC}); len(got) != 1 || got[0] != MetricTCC {
		t.Fatalf("closure(tcc) = %v", got)
	}
}
