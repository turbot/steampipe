import { classNames } from "./styles";

const getIconClasses = (type) => {
  const coloredClasses = "text-3xl";
  switch (type) {
    case "alert":
      return classNames(coloredClasses, "text-alert");
    case "info":
      return classNames(coloredClasses, "text-info");
    case "ok":
      return classNames(coloredClasses, "text-ok");
    case "severity":
      return classNames(coloredClasses, "text-severity");
    default:
      return "text-black-scale-4 text-3xl";
  }
};

const getIconForType = (type, icon) => {
  if (!type && !icon) {
    return null;
  }

  if (icon) {
    return icon;
  }

  switch (type) {
    case "alert":
      return "materialsymbols-solid:error";
    case "ok":
      return "materialsymbols-solid:check_circle";
    case "info":
      return "materialsymbols-solid:info";
    case "severity":
      return "materialsymbols-solid:warning";
    default:
      return null;
  }
};

const getTextClasses = (type) => {
  switch (type) {
    case "alert":
      return "text-alert";
    case "info":
      return "text-info";
    case "ok":
      return "text-ok";
    case "severity":
      return "text-severity";
    default:
      return null;
  }
};

const getWrapperClasses = (type) => {
  switch (type) {
    case "alert":
      return "bg-dashboard-panel border border-alert";
    case "info":
      return "bg-dashboard-panel border border-info";
    case "ok":
      return "bg-dashboard-panel border border-ok ";
    case "severity":
      return "bg-dashboard-panel border border-severity";
    default:
      return "bg-dashboard-panel shadow-sm border border-gray-400";
  }
};

export { getIconClasses, getIconForType, getTextClasses, getWrapperClasses };
