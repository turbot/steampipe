import useDashboardIcons from "../../hooks/useDashboardIcons";

interface IconProps {
  className?: string;
  icon: string;
  style?: any;
  title?: string;
}

const Icon = ({ className = "h-6 w-6", icon, style, title }: IconProps) => {
  const icons = useDashboardIcons();
  let MatchingIcon = icons.materialSymbols[icon];

  if (MatchingIcon) {
    return (
      <MatchingIcon
        className={className}
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

  return <MatchingIcon className={className} style={style} title={title} />;
};

export default Icon;
