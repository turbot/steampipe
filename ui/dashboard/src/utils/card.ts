const getIconClasses = (type) => {
  switch (type) {
    case "info":
    case "ok":
    case "alert":
    case "severity":
      return "text-white opacity-40 text-3xl";
    default:
      return "text-black-scale-4 text-3xl";
  }
};

const getTextClasses = (type) => {
  switch (type) {
    case "alert":
      return "text-alert-inverse";
    case "info":
      return "text-info-inverse";
    case "ok":
      return "text-ok-inverse";
    case "severity":
      return "text-white";
    default:
      return null;
  }
};

const getWrapperClasses = (type) => {
  switch (type) {
    case "alert":
      return "bg-alert";
    case "info":
      return "bg-info";
    case "ok":
      return "bg-ok";
    case "severity":
      return "bg-yellow";
    default:
      return "bg-dashboard-panel print:bg-white shadow-sm print:shadow-none print:border print:border-gray-100";
  }
};

export { getIconClasses, getTextClasses, getWrapperClasses };
