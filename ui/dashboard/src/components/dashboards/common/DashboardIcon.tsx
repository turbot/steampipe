import Icon from "../../Icon";
import { classNames } from "../../../utils/styles";
import { useMemo } from "react";

interface DashboardIconProps {
  className?: string;
  icon?: string | null;
  style?: any;
  title?: string;
}

interface DashboardTextIconProps extends DashboardIconProps {
  icon: string;
}

interface DashboardImageIconProps extends DashboardIconProps {
  icon: string;
}

const useDashboardIconType = (icon) =>
  useMemo(() => {
    if (!icon) {
      return null;
    }

    // This gets parsed as a URL if we don't check first
    if (
      icon.startsWith("heroicons-outline:") ||
      icon.startsWith("heroicons-solid:") ||
      icon.startsWith("materialsymbols-outline:") ||
      icon.startsWith("materialsymbols-solid:")
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

const DashboardTextIcon = ({
  className,
  icon,
  style,
  title,
}: DashboardTextIconProps) => {
  const text = useMemo(() => {
    if (!icon) {
      return "";
    } else {
      return icon.substring(5);
    }
  }, [icon]);
  return (
    <div className={classNames(className, "flex items-center justify-center")}>
      <span style={style} title={title || text}>
        {text}
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
        <Icon className={className} icon={icon} style={style} title={title} />
      );
    case "text":
      return (
        <DashboardTextIcon
          className={className}
          icon={icon}
          style={style}
          title={title}
        />
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

export default DashboardIcon;

export { useDashboardIconType };
