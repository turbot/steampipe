import { DashboardActions, DashboardDataModeCLISnapshot } from "../../types";
import { LATEST_EXECUTION_SCHEMA_VERSION } from "../../constants/versions";
import { useDashboard } from "../../hooks/useDashboard";
import { useNavigate } from "react-router-dom";
import { useRef } from "react";

const migrateSnapshotFileToExecutionCompleteEvent = (snapshot) => {
  switch (snapshot.schema_version) {
    default:
      const {
        layout,
        panels,
        inputs,
        variables,
        search_path,
        start_time,
        end_time,
      } = snapshot;
      return {
        action: DashboardActions.EXECUTION_COMPLETE,
        schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
        snapshot: {
          schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
          layout,
          panels,
          inputs,
          variables,
          search_path,
          start_time,
          end_time,
        },
      };
  }
};

const OpenSnapshotButton = () => {
  const { dispatch } = useDashboard();
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const navigate = useNavigate();

  return (
    <>
      <span
        className="text-base text-foreground-lighter hover:text-foreground cursor-pointer"
        onClick={() => {
          fileInputRef.current?.click();
        }}
      >
        Open snapshot…
      </span>
      <input
        ref={fileInputRef}
        accept=".sps"
        className="hidden"
        id="open-snapshot"
        name="open-snapshot"
        type="file"
        onChange={(e) => {
          const files = e.target.files;
          if (!files || files.length === 0) {
            return;
          }
          const fileName = files[0].name;
          const fr = new FileReader();
          fr.onload = () => {
            if (!fr.result) {
              return;
            }

            e.target.value = "";
            try {
              const data = JSON.parse(fr.result.toString());
              const event = migrateSnapshotFileToExecutionCompleteEvent(data);
              dispatch({
                type: DashboardActions.CLEAR_DASHBOARD_INPUTS,
                recordInputsHistory: false,
              });
              dispatch({
                type: DashboardActions.SELECT_DASHBOARD,
                dashboard: null,
                recordInputsHistory: false,
              });
              navigate(`/snapshot/${fileName}`);
              dispatch({
                type: DashboardActions.SET_DATA_MODE,
                dataMode: DashboardDataModeCLISnapshot,
                snapshotFileName: fileName,
              });
              dispatch({
                type: DashboardActions.EXECUTION_COMPLETE,
                ...event,
              });
              dispatch({
                type: DashboardActions.SET_DASHBOARD_INPUTS,
                value: event.snapshot.inputs,
                recordInputsHistory: false,
              });
            } catch (err: any) {
              dispatch({
                type: DashboardActions.WORKSPACE_ERROR,
                error: "Unable to load snapshot:" + err.message,
              });
            }
          };
          fr.readAsText(files[0]);
        }}
      />
    </>
  );
};

export default OpenSnapshotButton;
