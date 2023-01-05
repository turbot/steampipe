import ErrorMessage from "../../../ErrorMessage";
import { usePanel } from "../../../../hooks/usePanel";

type PanelErrorProps = {
  error: string;
};

type PanelStatusProps = {
  showPanelError: boolean;
};

const PanelInitialized = () => {
  return (
    <div
      className={
        "flex w-full h-full p-2 break-keep bg-red-50 border-red-700 border text-red-700 justify-center items-center shadow rounded-md"
      }
    >
      <ErrorMessage error="Initialized" />
    </div>
  );
};

const PanelBlocked = () => {
  return (
    <div
      className={
        "flex w-full h-full p-2 break-keep bg-red-50 border-red-700 border text-red-700 justify-center items-center shadow rounded-md"
      }
    >
      <ErrorMessage error="Blocked" />
    </div>
  );
};

const PanelRunning = () => {
  return (
    <div
      className={
        "flex w-full h-full p-2 break-keep bg-red-50 border-red-700 border text-red-700 justify-center items-center shadow rounded-md"
      }
    >
      <ErrorMessage error="Running" />
    </div>
  );
};

const PanelCancelled = () => {
  return (
    <div
      className={
        "flex w-full h-full p-2 break-keep bg-red-50 border-red-700 border text-red-700 justify-center items-center shadow rounded-md"
      }
    >
      <ErrorMessage error="Cancelled" />
    </div>
  );
};

const PanelError = ({ error }: PanelErrorProps) => {
  return (
    <div
      className={
        "flex w-full h-full p-2 break-keep bg-red-50 border-red-700 border text-red-700 justify-center items-center shadow rounded-md"
      }
    >
      <ErrorMessage error={error} />
    </div>
  );
};

const PanelStatus = ({ showPanelError }: PanelStatusProps) => {
  const { definition } = usePanel();

  return (
    <>
      {definition.status === "initialized" && <PanelInitialized />}
      {definition.status === "blocked" && <PanelBlocked />}
      {definition.status === "running" && <PanelRunning />}
      {definition.status === "cancelled" && <PanelCancelled />}
      {definition.status === "error" &&
        !!definition.error &&
        showPanelError && <PanelError error={definition.error as string} />}
    </>
  );
};

export default PanelStatus;
