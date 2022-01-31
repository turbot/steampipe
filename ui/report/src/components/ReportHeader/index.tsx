import { Link } from "react-router-dom";

const ReportHeader = ({ children }) => (
  <div className="w-screen print:hidden">
    <div className="flex items-center p-4 py-3 bg-background border-b border-black-scale-3 overflow-y-scroll">
      <Link to="/">
        <img
          src="/favicon.svg"
          alt="Steampipe Logo"
          style={{ height: "28px", paddingLeft: "3px" }}
        />
      </Link>
      <div className="flex-1 pl-4 sm:pl-6">{children}</div>
    </div>
  </div>
);

export default ReportHeader;
