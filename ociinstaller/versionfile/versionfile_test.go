package versionfile

import (
	"os"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {

	var v PluginVersionFile
	var vDb DBVersionFile

	fileName := "test.json"
	timeNow := time.Now()

	vDb.EmbeddedDB.Version = "0.0.1"
	vDb.EmbeddedDB.Name = "embeddedDb"
	vDb.EmbeddedDB.ImageDigest = "111111111111"
	vDb.EmbeddedDB.InstalledFrom = "hub.steampipe.io/core/embedded-postgres:latest"
	vDb.EmbeddedDB.LastCheckedDate = timeNow.Format(time.UnixDate)
	vDb.EmbeddedDB.InstallDate = timeNow.Format(time.UnixDate)

	timeNow2 := timeNow.Add(time.Minute * 10)

	vDb.FdwExtension.Version = "1.0.1"
	vDb.FdwExtension.Name = "fdwExtension"
	vDb.FdwExtension.ImageDigest = "2222222222"
	vDb.FdwExtension.InstalledFrom = "hub.steampipe.io/core/hub-extension:latest"
	vDb.FdwExtension.LastCheckedDate = timeNow2.Format(time.UnixDate)
	vDb.FdwExtension.InstallDate = timeNow2.Format(time.UnixDate)

	v.Plugins = make(map[string]*(InstalledVersion))

	awsPlugin := InstalledVersion{
		Name:            "hub.steampipe.io/steampipe/plugin/turbot/aws@latest",
		Version:         "0.0.3",
		ImageDigest:     "88995cc15963225884b825b12409f798b24aa7364bbf35a83d3a5fb5db85f346",
		InstalledFrom:   "hub.steampipe.io/steampipe/plugin/turbot/aws:latest",
		LastCheckedDate: timeNow2.Format(time.UnixDate),
		InstallDate:     timeNow2.Format(time.UnixDate),
	}

	v.Plugins[awsPlugin.Name] = &awsPlugin

	googlePlugin := InstalledVersion{
		Name:            "hub.steampipe.io/steampipe/plugin/turbot/google@1",
		Version:         "1.0.7",
		ImageDigest:     "3211232123654987313216549876516351",
		InstalledFrom:   "hub.steampipe.io/steampipe/plugin/turbot/google:1",
		LastCheckedDate: timeNow2.Format(time.UnixDate),
		InstallDate:     timeNow2.Format(time.UnixDate),
	}
	v.Plugins[googlePlugin.Name] = &googlePlugin

	//v.Plugins = append(v.Plugins, googlePlugin)

	if err := vDb.writeForDB(fileName); err != nil {
		t.Errorf("\nError writing file: %s", err.Error())
	}

	v2, err := readforDB(fileName)
	if err != nil {
		t.Errorf("\nError reading file: %s", err.Error())
	}
	v3, err := readForPlugin(fileName)
	if err != nil {
		t.Errorf("\nError reading file: %s", err.Error())
	}

	if v2.EmbeddedDB.Version != vDb.EmbeddedDB.Version {
		t.Errorf("\nError EmbeddedDB.Version is: %s, expected %s", v2.EmbeddedDB.Version, vDb.EmbeddedDB.Version)
	}
	if v2.EmbeddedDB.Name != vDb.EmbeddedDB.Name {
		t.Errorf("\nError EmbeddedDB.Name is: %s, expected %s", v2.EmbeddedDB.Name, vDb.EmbeddedDB.Name)
	}
	if v2.EmbeddedDB.ImageDigest != vDb.EmbeddedDB.ImageDigest {
		t.Errorf("\nError EmbeddedDB.ImageDigest is: %s, expected %s", v2.EmbeddedDB.ImageDigest, vDb.EmbeddedDB.ImageDigest)
	}
	if v2.EmbeddedDB.InstalledFrom != vDb.EmbeddedDB.InstalledFrom {
		t.Errorf("\nError EmbeddedDB.InstalledFrom is: %s, expected %s", v2.EmbeddedDB.InstalledFrom, vDb.EmbeddedDB.InstalledFrom)
	}
	if v2.EmbeddedDB.LastCheckedDate != vDb.EmbeddedDB.LastCheckedDate {
		t.Errorf("\nError EmbeddedDB.LastCheckedDate is: %s, expected %s", v2.EmbeddedDB.LastCheckedDate, vDb.EmbeddedDB.LastCheckedDate)
	}
	if v2.EmbeddedDB.InstallDate != vDb.EmbeddedDB.InstallDate {
		t.Errorf("\nError EmbeddedDB.InstallDate is: %s, expected %s", v2.EmbeddedDB.InstallDate, vDb.EmbeddedDB.InstallDate)
	}

	if len(v3.Plugins) != 2 {
		t.Errorf("\nexpected 2 plugins, found %d", len(v3.Plugins))
	}

	os.Remove(fileName)

}
