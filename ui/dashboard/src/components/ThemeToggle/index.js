import Icon from "../Icon";
import { ThemeNames } from "../../hooks/useTheme";
import { useDashboard } from "../../hooks/useDashboard";

const ThemeToggle = () => {
  const {
    themeContext: { setTheme, theme },
  } = useDashboard();
  return (
    <button
      type="button"
      className={
        (theme.name === ThemeNames.STEAMPIPE_DEFAULT
          ? "bg-gray-200"
          : "bg-gray-500") +
        " relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
      }
      onClick={() =>
        setTheme(
          theme.name === ThemeNames.STEAMPIPE_DEFAULT
            ? ThemeNames.STEAMPIPE_DARK
            : ThemeNames.STEAMPIPE_DEFAULT
        )
      }
      aria-pressed="false"
    >
      <span className="sr-only">Use setting</span>
      <span
        className={
          (theme.name === ThemeNames.STEAMPIPE_DEFAULT
            ? "translate-x-0"
            : "translate-x-5") +
          " pointer-events-none relative inline-block h-5 w-5 rounded-full bg-background-panel shadow transform ring-0 transition ease-in-out duration-200"
        }
      >
        <span
          className={
            (theme.name === ThemeNames.STEAMPIPE_DEFAULT
              ? "opacity-100 ease-in duration-200"
              : "opacity-0 ease-out duration-100") +
            " absolute inset-0 h-full w-full flex items-center justify-center transition-opacity text-gray-500"
          }
          aria-hidden="true"
        >
          <Icon icon="heroicons-solid:sun" />
        </span>
        <span
          className={
            (theme.name === ThemeNames.STEAMPIPE_DEFAULT
              ? "opacity-0 ease-out duration-100"
              : "opacity-100 ease-in duration-200") +
            " absolute inset-0 h-full w-full flex items-center justify-center transition-opacity text-gray-500"
          }
          aria-hidden="true"
        >
          <Icon icon="heroicons-solid:moon" />
        </span>
      </span>
    </button>
  );
};

export default ThemeToggle;
