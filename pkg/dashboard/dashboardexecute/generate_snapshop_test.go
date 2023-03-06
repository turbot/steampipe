package dashboardexecute

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"
)

func generateSnapshot(resource string, wg *sync.WaitGroup) {
	cmd := exec.Command(
		"steampipe",
		"dashboard",
		resource,
		"--snapshot",
		"--snapshot-location", "/Users/kai/snap",
		"--mod-location", "/Users/kai/Dev/github/turbot/steampipe-mod-aws-insights",
		"--cloud-host", "cloud.steampipe.io",
		"--output=none")

	cmd.Env = os.Environ()
	var stdErr bytes.Buffer
	var stdOut bytes.Buffer

	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut

	err := cmd.Run()

	if err != nil {
		fmt.Println(fmt.Sprintf("Dashboard resource %s error:", resource), stdErr.String())
	} else {
		fmt.Println(fmt.Sprintf("Dashboard resource %s succeeded:", resource), stdOut.String())
	}

	wg.Done()
}

func TestConcurrentSnapshots(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(3)
	go generateSnapshot("aws_insights.dashboard.iam_group_dashboard", &wg)
	go generateSnapshot("aws_insights.dashboard.iam_role_dashboard", &wg)
	go generateSnapshot("aws_insights.dashboard.iam_user_dashboard", &wg)
	wg.Wait()
}
