import { Link } from "react-router-dom";
// @ts-ignore
import { ReactComponent as Logo } from "./logos/steampipe-logo.svg";
// @ts-ignore
import { ReactComponent as LogoWordmarkColor } from "./logos/steampipe-logo-wordmark-color.svg";
// @ts-ignore
import { ReactComponent as LogoWordmarkDark } from "./logos/steampipe-logo-wordmark-darkmode.svg";
import { ThemeNames, useTheme } from "../../hooks/useTheme";

const SteampipeLogo = () => {
  const { theme } = useTheme();
  return (
    <div className="mr-4">
      <Link to="/">
        <div className="block md:hidden w-8">
          <Logo />
        </div>
        <div className="hidden md:block w-48">
          {theme.name === ThemeNames.STEAMPIPE_DEFAULT && <LogoWordmarkColor />}
          {theme.name === ThemeNames.STEAMPIPE_DARK && <LogoWordmarkDark />}
        </div>
      </Link>
    </div>
  );
};

export default SteampipeLogo;
