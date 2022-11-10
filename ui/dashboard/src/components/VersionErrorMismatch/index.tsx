import semver from "semver";

const VersionErrorMismatch = ({ cliVersion, uiVersion }) => {
  const uiOlder = semver.lt(uiVersion, cliVersion);
  return (
    <div className="space-y-2 text-black">
      <p>
        {!uiOlder && (
          <>Steampipe Dashboard UI version is newer than the CLI version.</>
        )}
        {uiOlder && (
          <>Steampipe Dashboard UI version is older than the CLI version.</>
        )}
      </p>
      <div>
        <span className="block">UI:</span>
        <span className="font-semibold">{uiVersion}</span>
      </div>
      <div>
        <span className="block">CLI:</span>
        <span className="font-semibold">{cliVersion}</span>
      </div>
      <p>
        {!uiOlder && (
          <>Please stop and restart your Steampipe dashboard process.</>
        )}
        {uiOlder && (
          <>Please hard refresh this page, or close and re-open your browser.</>
        )}
      </p>
    </div>
  );
};

export default VersionErrorMismatch;
