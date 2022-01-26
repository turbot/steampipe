import { Link } from "react-router-dom";
import { useReport } from "../../hooks/useReport";

const ReportList = () => {
  const { availableReportsLoaded, reports, selectedReport } = useReport();
  if (selectedReport) {
    return null;
  }

  return (
    <div className="w-full p-4">
      {!availableReportsLoaded && <></>}
      {availableReportsLoaded && (
        <>
          <h2 className="text-2xl font-medium">
            Welcome to the Steampipe Reporting UI
          </h2>
          {reports.length === 0 ? (
            <p className="py-2 italic">
              No reports defined. Please define at least one report in the mod.
            </p>
          ) : null}
          {reports.length > 0 ? (
            <p className="py-2">Please select a report from the list below:</p>
          ) : null}
          <ul className="list-none list-inside">
            {reports.map((availableReport) => (
              <li key={availableReport.name} className="pl-4">
                <Link
                  className="link-highlight"
                  to={`/${availableReport.name}`}
                >
                  {availableReport.title || availableReport.name}
                </Link>
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
};

export default ReportList;
