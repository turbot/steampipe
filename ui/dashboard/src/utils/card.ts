import { classNames } from "./styles";

const getIconClasses = (type) => {
  const coloredClasses = "text-3xl fill-white opacity-40 print:opacity-100";
  switch (type) {
    case "alert":
      return classNames(coloredClasses, "print:fill-alert");
    case "info":
      return classNames(coloredClasses, "print:fill-info");
    case "ok":
      return classNames(coloredClasses, "print:fill-ok");
    case "severity":
      return classNames(coloredClasses, "print:fill-yellow");
    default:
      return "fill-black-scale-4 text-3xl";
  }
};

const getTextClasses = (type) => {
  switch (type) {
    case "alert":
      return "text-alert-inverse print:text-foreground";
    case "info":
      return "text-info-inverse print:text-foreground";
    case "ok":
      return "text-ok-inverse print:text-foreground";
    case "severity":
      return "text-white print:text-foreground";
    default:
      return null;
  }
};

const getWrapperClasses = (type) => {
  switch (type) {
    case "alert":
      return "bg-alert border-alert print:border-2 print:bg-white";
    case "info":
      return "bg-info border-info print:border-2 print:bg-white";
    case "ok":
      return "bg-ok border-ok print:border-2 print:bg-white";
    case "severity":
      return "bg-yellow border-yellow print:border-2 print:bg-white";
    default:
      return "bg-dashboard-panel shadow-sm border-gray-400 print:border-2 print:shadow-none print:bg-white";
  }
};

export { getIconClasses, getTextClasses, getWrapperClasses };
