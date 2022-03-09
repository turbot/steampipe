import DashboardSearch from "../DashboardSearch";
import DashboardTagGroupSelect from "../DashboardTagGroupSelect";
import ExternalLink from "../ExternalLink";
import Logo from "./Logo";
import ThemeToggle from "../ThemeToggle";

const DashboardHeader = () => (
  <div className="flex w-screen px-4 py-3 items-center justify-between space-x-4 bg-background-panel border-b border-table-divide print:hidden">
    <Logo />
    <div className="w-96">
      <DashboardSearch />
    </div>
    <DashboardTagGroupSelect />
    <div className="w-full space-x-4 md:space-x-8 flex items-center justify-end">
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
