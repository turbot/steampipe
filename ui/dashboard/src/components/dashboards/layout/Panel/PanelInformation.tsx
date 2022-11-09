import { usePanel } from "../../../../hooks/usePanel";

const PanelInformation = () => {
  const { showPanelInformation, panelInformation } = usePanel();

  if (!showPanelInformation) {
    return null;
  }

  return (
    <div className="absolute h-[97%] overflow-y-scroll z-50 top-2 right-2 p-2 max-w-sm bg-dashboard-panel border border-divide rounded-md text-sm">
      {panelInformation}
    </div>
  );
};

export default PanelInformation;
