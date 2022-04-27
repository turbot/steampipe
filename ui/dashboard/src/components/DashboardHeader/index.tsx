import DashboardSearch from "../DashboardSearch";
import DashboardTagGroupSelect from "../DashboardTagGroupSelect";
import SteampipeLogo from "./SteampipeLogo";
import ThemeToggle from "../ThemeToggle";
import { classNames } from "../../utils/styles";
import { ThemeNames, useTheme } from "../../hooks/useTheme";
import { useDashboard } from "../../hooks/useDashboard";

const DashboardHeader = () => {
  const { theme } = useTheme();
  const {
    components: { ExternalLink },
  } = useDashboard();
  return (
    <div
      className={classNames(
        "flex w-screen px-4 py-3 items-center justify-between space-x-2 md:space-x-4 bg-dashboard-panel border-b print:hidden",
        theme.name === ThemeNames.STEAMPIPE_DARK
          ? "border-table-divide"
          : "border-background"
      )}
    >
      <SteampipeLogo />
      <div className="flex flex-grow space-x-2 md:space-x-4">
        <DashboardSearch />
        <DashboardTagGroupSelect />
      </div>
      <div className="space-x-2 sm:space-x-4 md:space-x-8 flex items-center justify-end">
        <ExternalLink
          className="text-base text-foreground-lighter hover:text-foreground"
          to="https://hub.steampipe.io"
          withReferrer={true}
        >
          <>Hub</>
        </ExternalLink>
        <ExternalLink
          className="text-base text-foreground-lighter hover:text-foreground"
          to="https://steampipe.io/docs"
          withReferrer={true}
        >
          <>Docs</>
        </ExternalLink>
        <ThemeToggle />
      </div>
    </div>
  );
};

export default DashboardHeader;
