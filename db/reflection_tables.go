package db

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/workspace"
)

func CreateReflectionTables(workspace *workspace.Workspace, client *Client) error {

	standardColumns := `  mod_name          varchar(40),
  resource_name     varchar(40),
  file_name         varchar(40),
  start_line_number integer,
  end_line_number   integer,
  auto_generated    boolean,          
  source_definition text,
  title             varchar(40),
  description       varchar(40),
  labels            varchar(40)[]`

	reflectionTables := map[string]string{
		"steampipe_query": fmt.Sprintf(`%s, 
  sql               text`, standardColumns),

		"steampipe_control": fmt.Sprintf(`%s, 
  query             text`, standardColumns),

		"steampipe_control_group": fmt.Sprintf(`%s, 
  parent            varchar(40)`, standardColumns),
	}

	var sql []string
	for tableName, columns := range reflectionTables {
		creationSql := fmt.Sprintf(`create temp table %s (
%s
);
`, tableName, columns)

		sql = append(sql, creationSql)
	}

	res, err := client.ExecuteSync(strings.Join(sql, "\n"))

	// now populate the tables
	//return populateReflectionTables(workspace, client)
	fmt.Println(res)
	fmt.Println(err)
	return nil
}
