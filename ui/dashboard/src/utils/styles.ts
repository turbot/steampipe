const classNames = (
  ...classes: (string | null | undefined)[]
): string | undefined => {
  const filtered = classes.filter(Boolean);
  return filtered.length > 0 ? filtered.join(" ") : undefined;
};

export { classNames };
