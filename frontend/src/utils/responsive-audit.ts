export const RESPONSIVE_ISSUES = [
  {
    file: "frontend/src/css/mobile.css",
    issues: [
      "Check all @media queries for consistency with breakpoints (736px, 480px, 400px)",
      "Ensure z-index values use CSS custom properties",
      "Verify fixed positioning elements don't overlap on small screens",
    ],
  },
  {
    file: "frontend/src/css/header-mobile.css",
    issues: [
      "Remove duplicate search hiding rules now handled by components",
      "Check header height adjustments on very small screens (<400px)",
      "Ensure action buttons have proper touch targets (min 44x44px)",
    ],
  },
  {
    file: "frontend/src/views/Layout.vue",
    issues: [
      "Verify sidebar overlay doesn't conflict with other overlays",
      "Check transition animations work on mobile",
      "Ensure mobile layout has proper viewport handling",
    ],
  },
  {
    file: "frontend/src/components/Sidebar.vue",
    issues: [
      "Check mobile sidebar collapse behavior",
      "Verify menu items are accessible on touch devices",
    ],
  },
  {
    file: "frontend/src/components/Breadcrumbs.vue",
    issues: [
      "Check truncation on mobile screens",
      "Verify touch targets for breadcrumb items",
    ],
  },
];

export const auditResponsiveIssues = () => {
  const issues: string[] = [];

  RESPONSIVE_ISSUES.forEach(({ file, issues: fileIssues }) => {
    console.log(`\n📋 ${file}:`);
    fileIssues.forEach((issue) => {
      console.log(`  • ${issue}`);
    });
  });

  return issues;
};
