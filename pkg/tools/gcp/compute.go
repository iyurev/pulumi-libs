package gcp

import (
	"errors"
	"fmt"
	"github.com/iyurev/pulumi-libs/pkg/constants/gcp"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	ErrWrongInstanceSize = errors.New("wrong instance size")
)

func machineTypeFromConfig(instSize string) (string, error) {
	switch instSize {
	case "Micro":
		return gcp.MachineTypeMedium, nil
	case "Small":
		return gcp.MachineTypeStandard2, nil
	case "Medium":
		return gcp.MachineTypeStandard4, nil
	case "Large":
		return gcp.MachineTypeStandard8, nil
	case "":
		return gcp.MachineTypeMedium, nil
	default:
		return "", ErrWrongInstanceSize
	}

}

type machineTypeBuilder struct {
	machineType string
	series      string
	zone        string
}

func (m *machineTypeBuilder) Done() pulumi.StringPtrInput {
	return pulumi.StringPtr(fmt.Sprintf("zones/%s/machineTypes/%s-%s", m.zone, m.series, m.machineType))
}

func NewMachineTypeBuilder(zone string) *machineTypeBuilder {
	return &machineTypeBuilder{zone: zone, machineType: gcp.MachineTypeMedium, series: gcp.MachineSeriesE2}
}

func (m *machineTypeBuilder) TypeMedium() *machineTypeBuilder {
	m.machineType = gcp.MachineTypeMedium
	return m
}

func (m *machineTypeBuilder) TypeStandard2() *machineTypeBuilder {
	m.machineType = gcp.MachineTypeStandard2
	return m
}

func (m *machineTypeBuilder) TypeStandard4() *machineTypeBuilder {
	m.machineType = gcp.MachineTypeStandard4
	return m
}

func (m *machineTypeBuilder) TypeStandard8() *machineTypeBuilder {
	m.machineType = gcp.MachineTypeStandard8
	return m
}

type diskTypeBuilder struct {
	diskType string
	project  string
	zone     string
}

func (d *diskTypeBuilder) Done() pulumi.StringPtrInput {
	return pulumi.Sprintf("projects/%s/zones/%s/diskTypes/%s", d.project, d.zone, d.diskType)
}

func (d *diskTypeBuilder) WithSSDDisk() *diskTypeBuilder {
	d.diskType = gcp.DiskTypeSSD
	return d
}

func (d *diskTypeBuilder) WithBalancedDisk() *diskTypeBuilder {
	d.diskType = gcp.DiskTypeBalanced
	return d
}

func (d *diskTypeBuilder) WithStandardDisk() *diskTypeBuilder {
	d.diskType = gcp.DiskTypeStandard
	return d
}

func NewDiskTypeBuilder(zone, project string) *diskTypeBuilder {
	return &diskTypeBuilder{diskType: gcp.DiskTypeBalanced, zone: zone, project: project}
}
