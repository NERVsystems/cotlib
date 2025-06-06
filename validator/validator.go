package validator

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

//go:embed schemas/** schemas/details/__chat.xsd schemas/details/__chatreceipt.xsd schemas/details/__geofence.xsd schemas/details/__group.xsd schemas/details/__serverdestination.xsd schemas/details/__video.xsd
var schemasFS embed.FS

//go:embed schemas/details/environment.xsd
var takDetailsEnvironmentXSD []byte

//go:embed schemas/details/fileshare.xsd
var takDetailsFileshareXSD []byte

//go:embed schemas/details/precisionlocation.xsd
var takDetailsPrecisionLocationXSD []byte

//go:embed schemas/details/takv.xsd
var takDetailsTakvXSD []byte

//go:embed schemas/details/mission.xsd
var takDetailsMissionXSD []byte

//go:embed schemas/details/shape.xsd
var takDetailsShapeXSD []byte

//go:embed schemas/details/color.xsd
var takDetailsColorXSD []byte

//go:embed schemas/details/hierarchy.xsd
var takDetailsHierarchyXSD []byte

//go:embed schemas/details/__geofence.xsd
var takDetailsGeofenceXSD []byte

//go:embed schemas/details/strokeColor.xsd
var takDetailsStrokeColorXSD []byte

//go:embed schemas/details/strokeWeight.xsd
var takDetailsStrokeWeightXSD []byte

//go:embed schemas/details/fillColor.xsd
var takDetailsFillColorXSD []byte

//go:embed schemas/details/height.xsd
var takDetailsHeightXSD []byte

//go:embed schemas/details/height_unit.xsd
var takDetailsHeightUnitXSD []byte

//go:embed schemas/details/bullseye.xsd
var takDetailsBullseyeXSD []byte

//go:embed schemas/details/routeinfo.xsd
var takDetailsRouteInfoXSD []byte

var (
	schemas map[string]*Schema
	once    sync.Once
	initErr error
)

// test hooks
var (
	mkTemp         = os.MkdirTemp
	writeSchemasFn = writeSchemas
)

func writeSchemas(dir string) error {
	return fs.WalkDir(schemasFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == "." {
				return nil
			}
			return os.MkdirAll(filepath.Join(dir, path), 0o750)
		}
		data, err := schemasFS.ReadFile(path)
		if err != nil {
			return err
		}
		dest := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o600)
	})
}

func initSchemas() {
	schemas = make(map[string]*Schema)

	tmpDir, err := mkTemp("", "cotlib-schemas")
	if err != nil {
		initErr = fmt.Errorf("create temp dir: %w", err)
		return
	}
	defer os.RemoveAll(tmpDir)
	if err := writeSchemasFn(tmpDir); err != nil {
		initErr = fmt.Errorf("write schemas: %w", err)
		return
	}

	err = fs.WalkDir(schemasFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".xsd" {
			return nil
		}
		sc, err := CompileFile(filepath.Join(tmpDir, path))
		if err != nil {
			// Skip schemas that fail to compile
			return nil
		}
		rel := strings.TrimPrefix(path, "schemas/")
		name := strings.TrimSuffix(rel, ".xsd")
		name = strings.ReplaceAll(name, string(os.PathSeparator), "-")
		name = strings.ReplaceAll(name, " ", "_")
		schemas[name] = sc
		if strings.HasPrefix(name, "details-") {
			schemas["tak-"+name] = sc
		}
		return nil
	})
	if err != nil {
		initErr = err
	}

	takDetailsEnvironment, err := Compile(takDetailsEnvironmentXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details environment schema: %w", err)
		return
	}
	schemas["tak-details-environment"] = takDetailsEnvironment

	takDetailsFileshare, err := Compile(takDetailsFileshareXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details fileshare schema: %w", err)
		return
	}
	schemas["tak-details-fileshare"] = takDetailsFileshare

	takDetailsPrecisionLocation, err := Compile(takDetailsPrecisionLocationXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details precisionlocation schema: %w", err)
		return
	}
	schemas["tak-details-precisionlocation"] = takDetailsPrecisionLocation

	takDetailsTakv, err := Compile(takDetailsTakvXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details takv schema: %w", err)
		return
	}
	schemas["tak-details-takv"] = takDetailsTakv

	takDetailsMission, err := Compile(takDetailsMissionXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details mission schema: %w", err)
		return
	}
	schemas["tak-details-mission"] = takDetailsMission

	takDetailsShape, err := Compile(takDetailsShapeXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details shape schema: %w", err)
		return
	}
	schemas["tak-details-shape"] = takDetailsShape

	takDetailsColor, err := Compile(takDetailsColorXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details color schema: %w", err)
		return
	}
	schemas["tak-details-color"] = takDetailsColor

	takDetailsHierarchy, err := Compile(takDetailsHierarchyXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details hierarchy schema: %w", err)
		return
	}
	schemas["tak-details-hierarchy"] = takDetailsHierarchy

	takDetailsGeofence, err := Compile(takDetailsGeofenceXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details __geofence schema: %w", err)
		return
	}
	schemas["tak-details-__geofence"] = takDetailsGeofence

	takDetailsStrokeColor, err := Compile(takDetailsStrokeColorXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details strokecolor schema: %w", err)
		return
	}
	schemas["tak-details-strokeColor"] = takDetailsStrokeColor

	takDetailsStrokeWeight, err := Compile(takDetailsStrokeWeightXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details strokeweight schema: %w", err)
		return
	}
	schemas["tak-details-strokeWeight"] = takDetailsStrokeWeight

	takDetailsFillColor, err := Compile(takDetailsFillColorXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details fillcolor schema: %w", err)
		return
	}
	schemas["tak-details-fillColor"] = takDetailsFillColor

	takDetailsHeight, err := Compile(takDetailsHeightXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details height schema: %w", err)
		return
	}
	schemas["tak-details-height"] = takDetailsHeight

	takDetailsHeightUnit, err := Compile(takDetailsHeightUnitXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details height_unit schema: %w", err)
		return
	}
	schemas["tak-details-height_unit"] = takDetailsHeightUnit

	takDetailsBullseye, err := Compile(takDetailsBullseyeXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details bullseye schema: %w", err)
		return
	}
	schemas["tak-details-bullseye"] = takDetailsBullseye

	takDetailsRouteInfo, err := Compile(takDetailsRouteInfoXSD)
	if err != nil {
		initErr = fmt.Errorf("compile TAK details routeinfo schema: %w", err)
		return
	}
	schemas["tak-details-routeinfo"] = takDetailsRouteInfo

	eventPoint, err := Compile(eventPointXSD)
	if err != nil {
		initErr = fmt.Errorf("compile event point schema: %w", err)
		return
	}
	schemas["event-point"] = eventPoint
}

// ValidateAgainstSchema validates XML against a named schema.
func ValidateAgainstSchema(name string, xml []byte) error {
	once.Do(initSchemas)
	if initErr != nil {
		return initErr
	}
	s, ok := schemas[name]
	if !ok {
		return fmt.Errorf("unknown schema %s", name)
	}
	return s.Validate(xml)
}

// ListAvailableSchemas returns a list of all available schema names.
func ListAvailableSchemas() []string {
	once.Do(initSchemas)
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	return names
}
