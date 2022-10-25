import NeutralButton from "../forms/NeutralButton";
import { DashboardActions } from "../../types";
import { useDashboard } from "../../hooks/useDashboard";
import { useNavigate } from "react-router-dom";
import { useRef } from "react";

const OpenSnapshotButton = () => {
  const { dispatch } = useDashboard();
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const navigate = useNavigate();

  return (
    <>
      <NeutralButton
        onClick={() => {
          fileInputRef.current?.click();
        }}
      >
        <>Open Snapshot</>
      </NeutralButton>
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
          const fr = new FileReader();
          fr.onload = () => {
            if (!fr.result) {
              return;
            }
            e.target.value = "";
            const data = JSON.parse(fr.result.toString());
            const { action, inputs, ...rest } = data;
            dispatch({
              type: DashboardActions.CLEAR_DASHBOARD_INPUTS,
              recordInputsHistory: false,
            });
            dispatch({
              type: DashboardActions.SELECT_DASHBOARD,
              dashboard: null,
              recordInputsHistory: false,
            });
            dispatch({
              type: DashboardActions.SET_DATA_MODE,
              dataMode: "cli_snapshot",
            });
            dispatch({
              type: DashboardActions.EXECUTION_COMPLETE,
              inputs: inputs || {},
              ...rest,
            });
            dispatch({
              type: DashboardActions.SET_DASHBOARD_INPUTS,
              value: inputs || {},
              recordInputsHistory: false,
            });
            navigate(`/${data.layout.name}`);
          };
          fr.readAsText(files[0]);
        }}
      />
    </>
  );
};

export default OpenSnapshotButton;
