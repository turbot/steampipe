// Needs to be a single unbroken string so that the tailwind content purging
// includes the class
// TODO try intermediate state - anything less than 6 is 6
const getResponsivePanelWidthClass = (width: number): string => {
  switch (width) {
    case 0:
      // Hide anything with no width
      return "hidden";
    case 1:
      return "md:col-span-6 lg:col-span-1";
    case 2:
      return "md:col-span-6 lg:col-span-2";
    case 3:
      return "md:col-span-6 lg:col-span-3";
    case 4:
      return "md:col-span-6 lg:col-span-4";
    case 5:
      return "md:col-span-6 lg:col-span-5";
    case 6:
      return "md:col-span-6";
    case 7:
      return "md:col-span-7";
    case 8:
      return "md:col-span-8";
    case 9:
      return "md:col-span-9";
    case 10:
      return "md:col-span-10";
    case 11:
      return "md:col-span-11";
    default:
      // 12 or anything else returns 12
      return "md:col-span-12";
  }
};

export { getResponsivePanelWidthClass };
