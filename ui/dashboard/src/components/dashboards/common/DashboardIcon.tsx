import Icon from "../../Icon";
import { classNames } from "../../../utils/styles";
import { memo, useMemo } from "react";

interface DashboardIconProps {
  className?: string;
  icon?: string | null;
  style?: any;
  title?: string;
}

interface DashboardHeroIconProps extends DashboardIconProps {
  icon: string;
  style?: any;
}

interface DashboardImageIconProps extends DashboardIconProps {
  icon: string;
  style?: any;
}

const useDashboardIconType = (icon) =>
  useMemo(() => {
    if (!icon) {
      return null;
    }

    // This gets parsed as a URL if we don't check first
    if (
      icon.startsWith("heroicons-outline:") ||
      icon.startsWith("heroicons-solid:")
    ) {
      return "icon";
    }

    // Same for text - this gets parsed as a URL if we don't check first
    if (icon.startsWith("text:")) {
      return "text";
    }

    // If it looks like a URL, treat it like a URL
    try {
      new URL(icon);
      return "url";
    } catch {}

    // Else fall back to hero icons
    return "icon";
  }, [icon]);

const DashboardImageIcon = ({
  className,
  icon,
  style,
  title,
}: DashboardImageIconProps) => (
  <img className={className} src={icon} alt="" style={style} title={title} />
);

const DashboardHeroIcon = ({
  className,
  icon,
  style,
  title,
}: DashboardHeroIconProps) => (
  <Icon className={className} icon={icon} style={style} title={title} />
);

const DashboardTextIcon = ({ className, icon }: DashboardHeroIconProps) => {
  return (
    <div className={classNames(className, "flex items-center justify-center")}>
      <span className={"truncate"} title={icon.substring(5)}>
        {icon.substring(5)}
      </span>
    </div>
  );
};

const DashboardIcon = ({
  className,
  icon,
  style,
  title,
}: DashboardIconProps) => {
  // First work out the type of the provided icon
  const iconType = useDashboardIconType(icon);

  if (!icon || !iconType) {
    return null;
  }

  switch (iconType) {
    case "icon":
      return (
        <DashboardHeroIcon
          className={className}
          icon={icon}
          style={style}
          title={title}
        />
      );
    case "text":
      return (
        <DashboardTextIcon className={className} icon={icon} title={title} />
      );
    case "url":
      return (
        <DashboardImageIcon
          className={className}
          icon={icon}
          style={style}
          title={title}
        />
      );
    default:
      return null;
  }
};

export default memo(DashboardIcon);

export { useDashboardIconType };
