import DashboardSearch from "../DashboardSearch";
import DashboardTagGroupSelect from "../DashboardTagGroupSelect";
import ExternalLink from "../ExternalLink";
import SteampipeLogo from "./SteampipeLogo";
import ThemeToggle from "../ThemeToggle";

const DashboardHeader = () => (
  <div className="flex w-screen px-4 py-3 items-center justify-between space-x-2 md:space-x-4 bg-dashboard-panel border-b border-table-divide print:hidden">
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

export default DashboardHeader;
