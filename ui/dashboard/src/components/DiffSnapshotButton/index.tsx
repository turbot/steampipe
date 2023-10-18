import { DashboardActions, DashboardDataModeCLISnapshot } from "../../types";
import { SnapshotDataToExecutionCompleteSchemaMigrator } from "../../utils/schema";
import { useDashboard } from "../../hooks/useDashboard";
import { useNavigate } from "react-router-dom";
import { useRef } from "react";

const DiffSnapshotButton = () => {
  const { dataMode, dispatch } = useDashboard();
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const navigate = useNavigate();

  if (dataMode !== DashboardDataModeCLISnapshot) {
    return null;
  }

  return (
    <>
      <span
        className="text-base text-foreground-lighter hover:text-foreground cursor-pointer"
        onClick={() => {
          fileInputRef.current?.click();
        }}
      >
        Diff
      </span>
      <input
        ref={fileInputRef}
        accept=".sps"
        className="hidden"
        id="diff-snapshot"
        name="diff-snapshot"
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
              const eventMigrator =
                new SnapshotDataToExecutionCompleteSchemaMigrator();
              const migratedEvent = eventMigrator.toLatest(data);
              dispatch({
                type: DashboardActions.DIFF_SNAPSHOT,
                ...migratedEvent,
                snapshotFileName: fileName,
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

export default DiffSnapshotButton;
