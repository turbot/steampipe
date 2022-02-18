import Select from "react-select";
import { SelectOption } from "../dashboards/inputs/SelectInput";
import { ThemeNames, useTheme } from "../../hooks/useTheme";
import { useDashboard } from "../../hooks/useDashboard";
import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

const DashboardSelector = () => {
  const [, setRandomVal] = useState(0);
  const [value, setValue] = useState<SelectOption | null>(null);
  const { dashboards } = useDashboard();
  const { dashboardName } = useParams();
  const { theme, wrapperRef } = useTheme();
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
  }, [dashboardName]);

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

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  if (!wrapperRef) {
    return null;
  }

  // @ts-ignore
  const style = window.getComputedStyle(wrapperRef);
  const background = style.getPropertyValue("--color-background");
  const foreground = style.getPropertyValue("--color-foreground");
  const blackScale1 = style.getPropertyValue("--color-black-scale-1");
  const blackScale2 = style.getPropertyValue("--color-black-scale-2");
  const blackScale3 = style.getPropertyValue("--color-black-scale-3");

  const customStyles = {
    clearIndicator: (provided) => ({
      ...provided,
      cursor: "pointer",
    }),
    control: (provided, state) => {
      return {
        ...provided,
        backgroundColor:
          theme.name === ThemeNames.STEAMPIPE_DARK ? blackScale2 : background,
        borderColor: state.isFocused ? "#2684FF" : blackScale3,
        boxShadow: "none",
      };
    },
    dropdownIndicator: (provided) => ({
      ...provided,
      cursor: "pointer",
    }),
    singleValue: (provided) => ({
      ...provided,
      color: foreground,
    }),
    menu: (provided) => ({
      ...provided,
      backgroundColor:
        theme.name === ThemeNames.STEAMPIPE_DARK ? blackScale2 : background,
      border: `1px solid ${blackScale3}`,
      boxShadow: "none",
      marginTop: 0,
      marginBottom: 0,
    }),
    menuList: (provided) => ({
      ...provided,
      paddingTop: 0,
      paddingBottom: 0,
    }),
    option: (provided, state) => {
      return {
        ...provided,
        backgroundColor: state.isFocused ? blackScale1 : "none",
        color: foreground,
        overflow: "hidden",
        textOverflow: "ellipsis",
        whiteSpace: "nowrap",
      };
    },
  };

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
        styles={customStyles}
        value={value}
      />
    </form>
  );
};

export default DashboardSelector;
