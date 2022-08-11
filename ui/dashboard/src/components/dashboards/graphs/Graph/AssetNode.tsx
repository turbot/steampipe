import Properties from "./Properties";
import Tooltip from "./Tooltip";
import { Handle } from "react-flow-renderer";
import { memo } from "react";

interface AssetNodeProps {
  data: {
    icon?: string;
    label: string;
    properties?: {};
  };
}

const AssetNode = ({ data }: AssetNodeProps) => {
  const icon = data.icon ? data.icon : null;

  // Notes:
  // * The Handle elements seem to be required to allow the connectors to work.
  return (
    <>
      {/*@ts-ignore*/}
      <Handle type="target" />
      {/*@ts-ignore*/}
      <Handle type="source" />
      <Tooltip
        overlay={<Properties properties={data.properties} />}
        title={data.label}
      >
        <div className="flex flex-col items-center cursor-auto">
          <div className="py-2 px-2 rounded-full w-[35px] h-[35px] text-sm leading-[35px] my-0 mx-auto border border-divide cursor-grab">
            {icon && <img className="max-w-full" src={icon} />}
          </div>
          <div className="text-[7px] mt-1 bg-dashboard-panel text-foreground min-w-[35px]">
            {data.label}
          </div>
        </div>
      </Tooltip>
    </>
  );
};

export default memo(AssetNode);
