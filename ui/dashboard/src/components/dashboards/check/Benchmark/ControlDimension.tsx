import { stringToColor } from "../../../../utils/color";

const ControlDimension = ({ dimensionKey, dimensionValue }) => (
  <span
    className="rounded-md text-xs"
    style={{ color: stringToColor(dimensionValue) }}
    title={`${dimensionKey} = ${dimensionValue}`}
  >
    {dimensionValue}
  </span>
);

export default ControlDimension;
