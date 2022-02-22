import Select from "react-select";
import useSelectInputStyles from "../dashboards/inputs/SelectInput/useSelectInputStyles";
import { SelectOption } from "../dashboards/inputs/SelectInput";
import { useDashboard } from "../../hooks/useDashboard";
import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

const DashboardSelector = () => {
  const [value, setValue] = useState<SelectOption | null>(null);
  const { dashboards } = useDashboard();
  const { dashboardName } = useParams();
  const navigate = useNavigate();

  const options: SelectOption[] = useMemo(() => {
    if (!dashboards) {
      return [];
    }

    return dashboards.map((dashboard) => ({
      label: dashboard.title || dashboard.short_name,
      value: dashboard.full_name,
    }));
  }, [dashboards]);

  useEffect(() => {
    if (!dashboardName && value) {
      setValue(null);
    }
  }, [dashboardName, value]);

  useEffect(() => {
    if (!dashboardName) {
      return;
    }

    // If we haven't got the data we need yet...
    if (!options || options.length === 0) {
      return;
    }

    const foundOption = options.find(
      (option) => option.value === dashboardName
    );

    setValue(foundOption || null);
  }, [dashboardName, options]);

  const styles = useSelectInputStyles();

  if (!styles) {
    return null;
  }

  return (
    <form>
      <Select
        aria-labelledby="dashboard-selector.input"
        className="max-w-sm"
        classNamePrefix="select"
        menuPortalTarget={document.body}
        inputId="dashboard-selector.input"
        isDisabled={!dashboards || dashboards.length === 0}
        isLoading={!dashboards || dashboards.length === 0}
        isClearable
        isRtl={false}
        isSearchable
        isMulti={false}
        name="dashboard-selector"
        onChange={(e) => {
          if (!e) {
            navigate("/");
            return;
          }
          const dashboardId = e.value;
          navigate(`/${dashboardId}`);
        }}
        options={options}
        placeholder="Choose a dashboard..."
        styles={styles}
        value={value}
      />
    </form>
  );
};

export default DashboardSelector;
