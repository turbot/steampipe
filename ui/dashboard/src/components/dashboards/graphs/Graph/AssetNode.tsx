import { Handle } from "react-flow-renderer";
import { memo } from "react";

const AssetNode = ({ data }) => {
  console.log(data);
  const icon = data.icon ? data.icon : "aws_s3_bucket";

  // Notes:
  // * The Handle elements seem to be required to allow the connectors to work.
  return (
    <>
      {/*@ts-ignore*/}
      <Handle type="target" />
      {/*@ts-ignore*/}
      <Handle type="source" />
      <div className="flex flex-col items-center cursor-auto">
        <div className="py-2 px-2 rounded-full w-[35px] h-[35px] text-sm leading-[35px] my-0 mx-auto border border-divide cursor-grab">
          <img className="max-w-full" src={icon} />
        </div>
        <div className="text-[7px] mt-1 bg-dashboard-panel text-foreground min-w-[35px]">
          {data.label}
        </div>
      </div>
    </>
  );
};

export default memo(AssetNode);
