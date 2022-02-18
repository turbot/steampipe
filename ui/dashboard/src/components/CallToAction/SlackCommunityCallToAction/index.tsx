import CallToAction from "../index";
import ExternalLink from "../../ExternalLink";
import { FeedbackIcon, HeartIcon } from "../../../constants/icons";

const SlackCommunityCallToAction = ({}) => (
  <CallToAction
    icon={
      <div className="w-12 sm:w-16">
        <FeedbackIcon />
        {/*<Icon className="text-4xl sm:text-6xl" icon={feedbackIcon} />*/}
      </div>
    }
    title={
      <>
        We <HeartIcon className="inline w-6 h-6 -mt-1" /> feedback!
      </>
    }
    message={
      <>
        <p>Found a bug? Got an idea? Think something is great?</p>
      </>
    }
    action={
      <ExternalLink
        className="link-highlight"
        url="https://steampipe.slack.com/archives/C02NMCM2QKE"
        withReferrer={true}
      >
        Let us know in #cloud on Slack!
      </ExternalLink>
    }
  />
);

export default SlackCommunityCallToAction;
