package display

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/karrick/gows"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
)

// ShowPaged displays the `content` in a system dependent pager
func ShowPaged(ctx context.Context, filename string) {
	nixPager(ctx, filename)
	//if isPagerNeeded(content) && (runtime.GOOS == "darwin" || runtime.GOOS == "linux") {
	//	nixPager(ctx, content)
	//} else {
	//	nullPager(content)
	//}
}

func isPagerNeeded(content string) bool {
	// only show pager in interactive mode
	if !viper.GetBool(constants.ConfigKeyInteractive) {
		return false
	}

	maxCols, maxRow, _ := gows.GetWinSize()

	// let's scan through it instead of iterating over it fully
	sc := bufio.NewScanner(strings.NewReader(content))

	// explicitly allocate a large bugger for the scanner to use - otherwise we may fail for large rows
	buffSize := 256 * 1024
	buff := make([]byte, buffSize)
	sc.Buffer(buff, buffSize)

	lineCount := 0
	for sc.Scan() {
		line := sc.Text()
		lineCount++
		if lineCount > maxRow {
			return true
		}
		if len(line) > maxCols {
			return true
		}
	}
	return false
}

func nullPager(content string) {
	// just dump the whole thing out
	// we will use this for non-tty environments as well
	fmt.Print(content)
}

func nixPager(ctx context.Context, filename string) {
	cmdStr := fmt.Sprintf("cat %s | less -SRXF", filename)
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Dir = filepath.Dir(filename)
	execPager(ctx, cmd)

	//if isLessAvailable() {
	//	execPager(ctx, exec.Command("less", fmt.Sprintf("+F %s", filename), "-SRXF"))
	//} else if isMoreAvailable() {
	//	execPager(ctx, exec.Command("more"), content)
	//} else {
	//	nullPager(content)
	//}
}

func isLessAvailable() bool {
	_, err := exec.LookPath("less")
	return err == nil
}

func isMoreAvailable() bool {
	_, err := exec.LookPath("more")
	return err == nil
}

func execPager(ctx context.Context, cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	//cmd.Stdin = strings.NewReader(content)
	// run the command - it will block until the pager is exited
	err := cmd.Run()
	if err != nil {
		utils.ShowErrorWithMessage(ctx, err, "could not display results")
	}
}
