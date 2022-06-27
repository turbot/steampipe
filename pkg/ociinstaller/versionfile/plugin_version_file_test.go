package versionfile

import (
	"os"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {

	var v PluginVersionFile

	fileName := "test.json"
	timeNow := time.Now()
	timeNow2 := timeNow.Add(time.Minute * 10)
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
	if err := v.write(fileName); err != nil {
		t.Errorf("\nError writing file: %s", err.Error())
	}
	v2, err := readPluginVersionFile(fileName)
	if err != nil {
		t.Errorf("\nError reading file: %s", err.Error())
	}

	if len(v2.Plugins) != 2 {
		t.Errorf("\nexpected 2 plugins, found %d", len(v2.Plugins))
	}

	os.Remove(fileName)

}
