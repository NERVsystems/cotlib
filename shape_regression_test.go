package cotlib

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestEllipseWithoutLink ensures ellipse shapes work without <link> element.
// Regression test: ATAK accepts circles without link, cotlib should too.
func TestEllipseWithoutLink(t *testing.T) {
	now := time.Now().UTC()
	stale := now.Add(24 * time.Hour)
	nowStr := now.Format(time.RFC3339)
	staleStr := stale.Format(time.RFC3339)

	xml := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<event version='2.0' uid='test-no-link' type='u-r-b-c-c' time='%s' start='%s' stale='%s' how='h-e'>
	<point lat='1.0' lon='1.0' hae='0' ce='10' le='10'/>
	<detail>
		<shape>
			<ellipse major='500.0' minor='500.0' angle='360'/>
		</shape>
		<contact callsign='Test'/>
	</detail>
</event>`, nowStr, nowStr, staleStr)

	ctx := context.Background()
	evt, err := UnmarshalXMLEvent(ctx, []byte(xml))
	if err != nil {
		t.Fatalf("Ellipse without link should parse, got: %v", err)
	}

	if err := evt.Validate(); err != nil {
		t.Fatalf("Ellipse without link should validate, got: %v", err)
	}
}

// TestEllipseWithLink ensures ellipse shapes still work WITH <link> element.
func TestEllipseWithLink(t *testing.T) {
	now := time.Now().UTC()
	stale := now.Add(24 * time.Hour)
	nowStr := now.Format(time.RFC3339)
	staleStr := stale.Format(time.RFC3339)

	xml := fmt.Sprintf(`<?xml version='1.0' encoding='UTF-8'?>
<event version='2.0' uid='test-with-link' type='u-r-b-c-c' time='%s' start='%s' stale='%s' how='h-e'>
	<point lat='1.0' lon='1.0' hae='0' ce='10' le='10'/>
	<detail>
		<shape>
			<ellipse major='500.0' minor='500.0' angle='360'/>
			<link uid='test.Style' type='b-x-KmlStyle' relation='p-c'>
				<Style>
					<LineStyle>
						<color>ffff0000</color>
						<width>3.0</width>
					</LineStyle>
					<PolyStyle>
						<color>00ff0000</color>
					</PolyStyle>
				</Style>
			</link>
		</shape>
		<contact callsign='Test'/>
	</detail>
</event>`, nowStr, nowStr, staleStr)

	ctx := context.Background()
	evt, err := UnmarshalXMLEvent(ctx, []byte(xml))
	if err != nil {
		t.Fatalf("Ellipse with link should parse, got: %v", err)
	}

	if err := evt.Validate(); err != nil {
		t.Fatalf("Ellipse with link should validate, got: %v", err)
	}
}

// TestColorFlexibleAttributes ensures color works with value, argb, or both.
func TestColorFlexibleAttributes(t *testing.T) {
	now := time.Now().UTC()
	stale := now.Add(24 * time.Hour)
	nowStr := now.Format(time.RFC3339)
	staleStr := stale.Format(time.RFC3339)

	tests := []struct {
		name string
		xml  string
	}{
		{"only_argb", fmt.Sprintf(`<event version='2.0' uid='t1' type='a-f-G' time='%s' start='%s' stale='%s' how='h-e'><point lat='1.0' lon='1.0' hae='0' ce='10' le='10'/><detail><color argb='-1'/></detail></event>`, nowStr, nowStr, staleStr)},
		{"only_value", fmt.Sprintf(`<event version='2.0' uid='t2' type='a-f-G' time='%s' start='%s' stale='%s' how='h-e'><point lat='1.0' lon='1.0' hae='0' ce='10' le='10'/><detail><color value='-1'/></detail></event>`, nowStr, nowStr, staleStr)},
		{"both", fmt.Sprintf(`<event version='2.0' uid='t3' type='a-f-G' time='%s' start='%s' stale='%s' how='h-e'><point lat='1.0' lon='1.0' hae='0' ce='10' le='10'/><detail><color value='-1' argb='-1'/></detail></event>`, nowStr, nowStr, staleStr)},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt, err := UnmarshalXMLEvent(ctx, []byte(tt.xml))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if err := evt.Validate(); err != nil {
				t.Fatalf("Validate failed: %v", err)
			}
		})
	}
}
