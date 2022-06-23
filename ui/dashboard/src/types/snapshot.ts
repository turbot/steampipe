interface DashboardInputs {
  [name: string]: string;
}

interface DashboardVariables {
  [name: string]: any;
}

export interface DashboardSnapshot {
  id: string;
  dashboard_name: string;
  start_time: string;
  end_time: string;
  lineage: string;
  schema_version: string;
  search_path: string;
  variables: DashboardVariables;
  inputs: DashboardInputs;
}
