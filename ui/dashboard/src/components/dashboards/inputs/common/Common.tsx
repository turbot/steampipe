import { ColorGenerator } from "../../../../utils/color";
import { components, OptionProps, SingleValueProps } from "react-select";

const stringColorMap = {};
const colorGenerator = new ColorGenerator(24, 4);

const stringToColor = (str) => {
  if (stringColorMap[str]) {
    return stringColorMap[str];
  }
  const color = colorGenerator.nextColor().hex;
  stringColorMap[str] = color;
  return color;
};

const OptionTag = ({ tagKey, tagValue }) => (
  <span
    className="rounded-md text-xs"
    style={{ color: stringToColor(tagValue) }}
    title={`${tagKey} = ${tagValue}`}
  >
    {tagValue}
  </span>
);

const LabelTagWrapper = ({ label, tags }) => (
  <div className="space-x-2">
    {/*@ts-ignore*/}
    <span>{label}</span>
    {/*@ts-ignore*/}
    {Object.entries(tags || {}).map(([tagKey, tagValue]) => (
      <OptionTag key={tagKey} tagKey={tagKey} tagValue={tagValue} />
    ))}
  </div>
);

const OptionWithTags = (props: OptionProps) => (
  <components.Option {...props}>
    {/*@ts-ignore*/}
    <LabelTagWrapper label={props.data.label} tags={props.data.tags} />
  </components.Option>
);

const SingleValueWithTags = ({ children, ...props }: SingleValueProps) => {
  return (
    <components.SingleValue {...props}>
      {/*@ts-ignore*/}
      <LabelTagWrapper label={props.data.label} tags={props.data.tags} />
    </components.SingleValue>
  );
};

const MultiValueLabelWithTags = ({ children, ...props }: SingleValueProps) => {
  return (
    <components.MultiValueLabel {...props}>
      {/*@ts-ignore*/}
      <LabelTagWrapper label={props.data.label} tags={props.data.tags} />
    </components.MultiValueLabel>
  );
};

export { MultiValueLabelWithTags, OptionWithTags, SingleValueWithTags };
