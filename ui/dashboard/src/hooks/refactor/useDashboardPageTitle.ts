import { AvailableDashboard } from "../../types/dashboard";
import { useEffect } from "react";

const useDashboardPageTitle = (selectedDashboard?: AvailableDashboard) => {
  useEffect(() => {
    if (!selectedDashboard) {
      document.title = "Dashboards | Steampipe";
    } else {
      document.title = `${
        selectedDashboard.title || selectedDashboard.full_name
      } | Dashboards | Steampipe`;
    }
  }, [selectedDashboard]);
};

export default useDashboardPageTitle;
