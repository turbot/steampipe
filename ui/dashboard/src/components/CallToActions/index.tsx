import ExternalLink from "../ExternalLink";
import { classNames } from "../../utils/styles";
import {
  BuildDashboardIcon,
  CommunityIcon,
  InstallDashboardIcon,
} from "../../constants/icons";
import { ThemeNames, useTheme } from "../../hooks/useTheme";

const items = [
  {
    title: "Install a dashboard",
    description:
      "Steampipe Hub has hundreds of open source dashboards to get you started.",
    href: "https://hub.steampipe.io/mods?objectives=dashboard",
    icon: InstallDashboardIcon,
    background: "bg-steampipe-red",
    withReferrer: true,
  },
  {
    title: "Build a Dashboard",
    description:
      "It's easy to create your own dashboard as code! Start with this tutorial.",
    href: "https://steampipe.io/docs/mods/writing-dashboards",
    icon: BuildDashboardIcon,
    iconColor: "text-white",
    iconColorInverse: "text-black",
    background: "bg-black",
    backgroundInverse: "bg-white",
    withReferrer: true,
  },
  {
    title: "Join our Community",
    description:
      "Connect directly with Steampipe users and the development team in Slack.",
    href: "https://steampipe.io/community/join",
    icon: CommunityIcon,
    background: "bg-slack-aubergine",
    withReferrer: true,
  },
];

const CallToActions = () => {
  const { theme } = useTheme();
  return (
    <ul
      role="list"
      className="mt-6 border-t border-b border-table-divide py-6 space-y-6"
    >
      {items.map((item, itemIdx) => (
        <li key={itemIdx} className="flow-root">
          <div className="relative -m-2 p-2 flex items-center space-x-4 rounded-xl hover:bg-background-panel focus-within:ring-2 focus-within:ring-indigo-500">
            <div
              className={classNames(
                item.backgroundInverse &&
                  theme.name === ThemeNames.STEAMPIPE_DARK
                  ? item.backgroundInverse
                  : item.background,
                "flex-shrink-0 flex items-center justify-center h-16 w-16 rounded-lg"
              )}
            >
              <item.icon
                className={classNames(
                  "h-6 w-6",
                  item.iconColor || item.iconColorInverse
                    ? theme.name === ThemeNames.STEAMPIPE_DARK
                      ? item.iconColorInverse
                      : item.iconColor
                    : "text-white"
                )}
                aria-hidden="true"
              />
            </div>
            <div>
              <h3 className="text-sm font-medium text-foreground">
                <ExternalLink
                  to={item.href}
                  className="focus:outline-none"
                  withReferrer={item.withReferrer}
                >
                  <span className="absolute inset-0" aria-hidden="true" />
                  <>{item.title}</>
                  <span aria-hidden="true" className="ml-1">
                    &rarr;
                  </span>
                </ExternalLink>
              </h3>
              <p className="mt-1 text-sm text-foreground-light">
                {item.description}
              </p>
            </div>
          </div>
        </li>
      ))}
    </ul>
  );
};

export default CallToActions;
