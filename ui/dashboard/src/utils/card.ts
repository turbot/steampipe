import { classNames } from "./styles";

const getIconClasses = (type) => {
  const colouredClasses = "text-3xl text-white opacity-40 print:opacity-100";
  switch (type) {
    case "alert":
      return classNames(colouredClasses, "print:text-alert");
    case "info":
      return classNames(colouredClasses, "print:text-info");
    case "ok":
      return classNames(colouredClasses, "print:text-ok");
    case "severity":
      return classNames(colouredClasses, "print:text-yellow");
    default:
      return "text-black-scale-4 text-3xl";
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
      return "bg-dashboard-panel shadow-sm border-divide print:border-2 print:shadow-none print:bg-white";
  }
};

export { getIconClasses, getTextClasses, getWrapperClasses };
