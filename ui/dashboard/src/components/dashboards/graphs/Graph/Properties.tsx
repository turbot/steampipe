import { classNames } from "../../../../utils/styles";

const PropertyItemValue = ({ value }) => {
  if (value === null || value === undefined) {
    return (
      <span className="text-foreground-lightest">
        <>null</>
      </span>
    );
  }
  let renderValue: string = "";
  switch (typeof value) {
    case "object":
      renderValue = JSON.stringify(value, null, 2);
      break;
    default:
      renderValue = value.toString();
      break;
  }
  return (
    <span className={classNames("block", "break-words")}>{renderValue}</span>
  );
};

const PropertyItem = ({ name, value }) => {
  return (
    <div>
      <span className="block text-sm text-table-head truncate">{name}</span>
      <PropertyItemValue value={value} />
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
