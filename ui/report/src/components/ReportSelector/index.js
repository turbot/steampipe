import ThemeToggle from "../ThemeToggle";
import { useNavigate, useParams } from "react-router-dom";
import { useReport } from "../../hooks/useReport";

const ReportSelector = () => {
  const navigate = useNavigate();
  const { reports } = useReport();
  const { reportName } = useParams();
  return (
    <div className="flex justify-between items-center">
      <div className="flex items-center space-x-2">
        <select
          className="block w-full pl-3 pr-10 py-2 text-base bg-background border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
          id="report"
          name="report"
          onChange={(e) => {
            const reportId = e.target.value;
            navigate(`/${reportId}`);
          }}
          value={reportName || ""}
        >
          <option value="">Choose a report...</option>
          {reports.map((report) => (
            <option key={report.name} value={report.name}>
              {report.title || report.name}
            </option>
          ))}
        </select>
      </div>
      <ThemeToggle />
    </div>
  );
};

export default ReportSelector;
