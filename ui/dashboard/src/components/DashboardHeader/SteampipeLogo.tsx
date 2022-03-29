// @ts-ignore
import { ReactComponent as Logo } from "./logos/steampipe-logo.svg";
// @ts-ignore
import { ReactComponent as LogoWordmarkColor } from "./logos/steampipe-logo-wordmark-color.svg";
// @ts-ignore
import { ReactComponent as LogoWordmarkDark } from "./logos/steampipe-logo-wordmark-darkmode.svg";
import { ThemeNames } from "../../hooks/useTheme";
import { useDashboard } from "../../hooks/useDashboard";

const SteampipeLogo = () => {
  const {
    components: { ExternalLink },
    themeContext: { theme },
  } = useDashboard();

  return (
    <div className="mr-1 md:mr-4">
      <ExternalLink to="/">
        <div className="block md:hidden w-8">
          <Logo />
        </div>
        <div className="hidden md:block w-48">
          {theme.name === ThemeNames.STEAMPIPE_DEFAULT && <LogoWordmarkColor />}
          {theme.name === ThemeNames.STEAMPIPE_DARK && <LogoWordmarkDark />}
        </div>
      </ExternalLink>
    </div>
  );
};

export default SteampipeLogo;
