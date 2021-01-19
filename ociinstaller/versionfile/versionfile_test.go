package versionfile

import (
	"os"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {

	var v VersionFile

	fileName := "test.json"
	timeNow := time.Now()

	v.EmbeddedDB.Version = "0.0.1"
	v.EmbeddedDB.Name = "embeddedDb"
	v.EmbeddedDB.ImageDigest = "111111111111"
	v.EmbeddedDB.InstalledFrom = "hub.steampipe.io/core/embedded-postgres:latest"
	v.EmbeddedDB.LastCheckedDate = timeNow.Format(time.UnixDate)
	v.EmbeddedDB.InstallDate = timeNow.Format(time.UnixDate)

	timeNow2 := timeNow.Add(time.Minute * 10)

	v.FdwExtension.Version = "1.0.1"
	v.FdwExtension.Name = "fdwExtension"
	v.FdwExtension.ImageDigest = "2222222222"
	v.FdwExtension.InstalledFrom = "hub.steampipe.io/core/hub-extension:latest"
	v.FdwExtension.LastCheckedDate = timeNow2.Format(time.UnixDate)
	v.FdwExtension.InstallDate = timeNow2.Format(time.UnixDate)

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

	if err := v.write(fileName); err != nil {
		t.Errorf("\nError writing file: %s", err.Error())
	}

	v2, err := read(fileName)
	if err != nil {
		t.Errorf("\nError reading file: %s", err.Error())
	}

	if v2.EmbeddedDB.Version != v.EmbeddedDB.Version {
		t.Errorf("\nError EmbeddedDB.Version is: %s, expected %s", v2.EmbeddedDB.Version, v.EmbeddedDB.Version)
	}
	if v2.EmbeddedDB.Name != v.EmbeddedDB.Name {
		t.Errorf("\nError EmbeddedDB.Name is: %s, expected %s", v2.EmbeddedDB.Name, v.EmbeddedDB.Name)
	}
	if v2.EmbeddedDB.ImageDigest != v.EmbeddedDB.ImageDigest {
		t.Errorf("\nError EmbeddedDB.ImageDigest is: %s, expected %s", v2.EmbeddedDB.ImageDigest, v.EmbeddedDB.ImageDigest)
	}
	if v2.EmbeddedDB.InstalledFrom != v.EmbeddedDB.InstalledFrom {
		t.Errorf("\nError EmbeddedDB.InstalledFrom is: %s, expected %s", v2.EmbeddedDB.InstalledFrom, v.EmbeddedDB.InstalledFrom)
	}
	if v2.EmbeddedDB.LastCheckedDate != v.EmbeddedDB.LastCheckedDate {
		t.Errorf("\nError EmbeddedDB.LastCheckedDate is: %s, expected %s", v2.EmbeddedDB.LastCheckedDate, v.EmbeddedDB.LastCheckedDate)
	}
	if v2.EmbeddedDB.InstallDate != v.EmbeddedDB.InstallDate {
		t.Errorf("\nError EmbeddedDB.InstallDate is: %s, expected %s", v2.EmbeddedDB.InstallDate, v.EmbeddedDB.InstallDate)
	}

	if len(v2.Plugins) != 2 {
		t.Errorf("\nexpected 2 plugins, found %d", len(v2.Plugins))
	}

	os.Remove(fileName)

}
