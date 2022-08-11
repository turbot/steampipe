import { classNames } from "../../../../utils/styles";

const PropertyItem = ({ name, value }) => {
  return (
    <div>
      <span className="block text-sm text-table-head truncate">{name}</span>
      {value === null && (
        <span className="text-foreground-lightest">
          <>null</>
        </span>
      )}
      {value !== null && (
        <span className={classNames("block", "break-words")}>{value}</span>
      )}
    </div>
    // <div className="space-x-2">
    //   <span>{name}</span>
    //   <span>=</span>
    //   <span>{value}</span>
    // </div>
  );
};

const Properties = ({ properties = {} }) => {
  return (
    <div className="space-y-2">
      {Object.entries(properties || {}).map(([key, value]) => (
        <PropertyItem key={key} name={key} value={value} />
      ))}
    </div>
  );
};

export default Properties;
