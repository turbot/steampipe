import { Link } from "react-router-dom";
// @ts-ignore
import { ReactComponent as LogoColor } from "./steampipe-logo-wordmark-color.svg";
// @ts-ignore
import { ReactComponent as LogoDark } from "./steampipe-logo-wordmark-darkmode.svg";
import { ThemeNames, useTheme } from "../../hooks/useTheme";

const Logo = () => {
  const { theme } = useTheme();
  return (
    <div className="min-w-96 mr-4">
      <Link to="/">
        {theme.name === ThemeNames.STEAMPIPE_DEFAULT && <LogoColor />}
        {theme.name === ThemeNames.STEAMPIPE_DARK && <LogoDark />}
      </Link>
    </div>
  );
};

export default Logo;
