import ErrorMessage from "../../../ErrorMessage";
import { classNames } from "../../../../utils/styles";
import { PanelDefinition } from "../../../../types";
import { ReactNode } from "react";
import { usePanel } from "../../../../hooks/usePanel";

type PanelStatusProps = PanelStatusBaseProps & {
  definition: PanelDefinition;
  showPanelError: boolean;
};

type PanelStatusBaseProps = {
  definition: PanelDefinition;
};

type PanelErrorProps = PanelStatusBaseProps & {
  error: string;
};

type BasePanelStatusProps = {
  children: ReactNode;
  className?: string;
  definition: PanelDefinition;
};

const BasePanelStatus = ({
  children,
  className,
  definition,
}: BasePanelStatusProps) => (
  <div
    className={classNames(
      className,
      "flex w-full h-full p-2 break-keep border justify-center items-center shadow rounded-md",
      !!definition.title ? "rounded-t-none" : null
    )}
  >
    {children}
  </div>
);

const PanelInitialized = ({ definition }: PanelStatusBaseProps) => {
  return <BasePanelStatus definition={definition}>Initialized</BasePanelStatus>;
};

const PanelBlocked = ({ definition }) => {
  const { dependenciesByStatus } = usePanel();
  console.log(definition.name, dependenciesByStatus);
  return <BasePanelStatus definition={definition}>Blocked</BasePanelStatus>;
};

const PanelRunning = ({ definition }) => {
  return <BasePanelStatus definition={definition}>Running</BasePanelStatus>;
};

const PanelCancelled = ({ definition }) => {
  return <BasePanelStatus definition={definition}>Cancelled</BasePanelStatus>;
};

const PanelError = ({ definition, error }: PanelErrorProps) => {
  return (
    <BasePanelStatus
      className="bg-alert-light border-alert-light text-foreground"
      definition={definition}
    >
      <ErrorMessage error={error} />
    </BasePanelStatus>
  );
};

const PanelStatus = ({ definition, showPanelError }: PanelStatusProps) => (
  <>
    {definition.status === "initialized" && (
      <PanelInitialized definition={definition} />
    )}
    {definition.status === "blocked" && (
      <PanelBlocked definition={definition} />
    )}
    {definition.status === "running" && (
      <PanelRunning definition={definition} />
    )}
    {definition.status === "cancelled" && (
      <PanelCancelled definition={definition} />
    )}
    {definition.status === "error" && !!definition.error && showPanelError && (
      <PanelError definition={definition} error={definition.error as string} />
    )}
  </>
);

export default PanelStatus;
