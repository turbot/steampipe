import useDashboardIcons from "../../hooks/useDashboardIcons";

interface IconProps {
  className?: string;
  icon: string;
  style?: any;
  title?: string;
}

const Icon = ({ className = "h-6 w-6", icon, style, title }: IconProps) => {
  const icons = useDashboardIcons();
  let MatchingIcon;
  MatchingIcon = icons.materialSymbols[icon];
  if (MatchingIcon) {
    console.log("Using material symbol", icon);
  } else {
    MatchingIcon = icons.heroIcons[icon];
  }
  if (MatchingIcon) {
    console.log("Using hero icon", icon);
  } else {
    return null;
  }
  return (
    <MatchingIcon
      className={className}
      style={
        !!style
          ? { fill: style.color ? "currentColor" : undefined, ...style }
          : undefined
      }
      title={title}
      width="auto"
      height="auto"
    />
  );
};

export default Icon;
