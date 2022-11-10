import ExternalLink from "../ExternalLink";

const VersionErrorMismatch = ({ cliVersion, uiVersion }) => {
  return (
    <div className="space-y-2">
      <p>Steampipe Dashboard UI is running a different version to the CLI.</p>
      <div>
        <span className="block text-foreground-light">CLI:</span>
        <span className="font-semibold">{cliVersion}</span>
      </div>
      <div>
        <span className="block text-foreground-light">UI:</span>
        <span className="font-semibold">{uiVersion}</span>
      </div>
      <p>Please try the following:</p>
      <ul className="list-disc list-inside">
        <li>Stop and restart all Steampipe dashboard processes.</li>
        <li>Close and re-open your browser.</li>
      </ul>
      <p>
        If the issue persists, please let us know on our{" "}
        <ExternalLink
          className="link-highlight"
          to="https://steampipe.slack.com/archives/C01UECB59A7"
          target="_blank"
          withReferrer={false}
        >
          <>#steampipe</>
        </ExternalLink>{" "}
        Slack channel.
      </p>
    </div>
  );
};

export default VersionErrorMismatch;
