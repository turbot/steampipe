import { classNames } from "../../../../utils/styles";
import { usePanel } from "../../../../hooks/usePanel";

const PanelInformation = () => {
  const { showPanelInformation, panelInformation } = usePanel();

  if (!showPanelInformation) {
    return null;
  }

  return (
    <div
      className={classNames(
        "absolute h-full overflow-y-scroll z-50 top-0 right-0 p-3 max-w-sm bg-dashboard-panel border-l border-divide text-sm"
      )}
    >
      {panelInformation}
    </div>
  );
};

export default PanelInformation;
