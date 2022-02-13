import DashboardSelector from "../DashboardSelector";
import ThemeToggle from "../ThemeToggle";
import { Link } from "react-router-dom";

const DashboardHeader = () => (
  <div className="flex w-screen p-3 space-x-4 items-center bg-background border-b border-black-scale-3 print:hidden">
    <div>
      <Link to="/">
        <img
          src="/favicon.svg"
          alt="Steampipe Logo"
          style={{ height: "28px", paddingLeft: "3px" }}
        />
      </Link>
    </div>
    <div className="w-full grid grid-cols-12">
      <div className="col-span-10 md:col-span-6 lg:col-span-4">
        <DashboardSelector />
      </div>
      <div className="flex col-span-2 md:col-span-6 lg:col-span-8 items-center justify-end">
        <ThemeToggle />
      </div>
    </div>
  </div>
);

export default DashboardHeader;
