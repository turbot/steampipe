import * as outlineIconExports from "@heroicons/react/outline";
import * as solidIconExports from "@heroicons/react/solid";
import { kebabCase } from "lodash";

const icons = {};

const convertIconName = (name) => {
  let condensedName = name;
  const iconOccurrence = name.lastIndexOf("Icon");
  if (iconOccurrence >= 0) {
    condensedName = condensedName.substring(0, iconOccurrence);
  }
  return kebabCase(condensedName);
};

Object.entries(outlineIconExports).forEach(([name, exported]) => {
  const iconName = convertIconName(name);
  icons[iconName] = exported;
  icons[`heroicons-outline:${iconName}`] = exported;
});

Object.entries(solidIconExports).forEach(([name, exported]) => {
  const iconName = convertIconName(name);
  icons[`heroicons-solid:${iconName}`] = exported;
});

interface IconProps {
  className?: string;
  icon: string;
}

const Icon = ({ className = "h-6 w-6", icon }: IconProps) => {
  const MatchingIcon = icons[icon];
  if (!MatchingIcon) {
    return null;
  }
  return <MatchingIcon className={className} />;
};

export default Icon;
