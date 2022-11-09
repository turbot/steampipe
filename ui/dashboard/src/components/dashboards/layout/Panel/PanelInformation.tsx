import { usePanel } from "../../../../hooks/usePanel";

const PanelInformation = () => {
  const { showPanelInformation, panelInformation } = usePanel();

  if (!showPanelInformation) {
    return null;
  }

  return (
    <div className="absolute top-1 right-1 p-2 max-w-sm border border-divide rounded-md text-sm">
      {panelInformation}
    </div>
  );
};

export default PanelInformation;
