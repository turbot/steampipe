import useDashboardIcons from "../../hooks/useDashboardIcons";
import { noop } from "lodash";

type IconProps = {
  className?: string;
  icon: string;
  onClick?: () => void;
  style?: any;
  title?: string;
};

const Icon = ({
  className = "h-6 w-6",
  icon,
  onClick = noop,
  style,
  title,
}: IconProps) => {
  const icons = useDashboardIcons();
  let MatchingIcon = icons.materialSymbols[icon];

  if (MatchingIcon) {
    return (
      <MatchingIcon
        className={className}
        onClick={onClick}
        style={{
          fill: "currentColor",
          color: style ? style.color : undefined,
        }}
        title={title}
      />
    );
  } else {
    MatchingIcon = icons.heroIcons[icon];
  }

  if (!MatchingIcon) {
    return null;
  }

  return (
    <MatchingIcon
      className={className}
      onClick={onClick}
      style={style}
      title={title}
    />
  );
};

export default Icon;
