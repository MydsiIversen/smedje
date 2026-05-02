export function PrivacyPage() {
  return (
    <div className="min-h-screen bg-anvil text-foreground flex justify-center py-12 px-6">
      <div className="max-w-[720px] w-full">
        <a
          href="/"
          onClick={(e) => {
            e.preventDefault()
            window.location.pathname = "/"
          }}
          className="text-sm text-muted-foreground hover:text-foreground transition-colors duration-150"
        >
          &larr; Back to Smedje
        </a>

        <h1 className="font-mono text-lg text-foreground mt-6 mb-4">Privacy</h1>
        <p className="text-sm leading-relaxed text-foreground mb-4">
          Smedje is a local-first tool. Every key, ID, certificate, and password
          is generated in your browser or on your machine. Generated values never
          leave your device and are never transmitted to any server.
        </p>

        <h2 className="font-mono text-lg text-foreground mt-8 mb-3">Analytics</h2>
        <p className="text-sm leading-relaxed text-foreground mb-3">
          The public demo at{" "}
          <span className="font-mono bg-panel px-1">app.smedje.net</span>{" "}
          uses self-hosted Umami analytics at{" "}
          <span className="font-mono bg-panel px-1">analytics.smedje.net</span>{" "}
          to collect aggregate, anonymous usage data:
        </p>
        <ul className="text-sm leading-relaxed text-foreground mb-4 list-disc pl-6 space-y-1">
          <li>Page views (which pages are visited)</li>
          <li>Which generators are popular</li>
          <li>General usage patterns</li>
        </ul>
        <p className="text-sm leading-relaxed text-foreground mb-3">
          We do not collect:
        </p>
        <ul className="text-sm leading-relaxed text-foreground mb-4 list-disc pl-6 space-y-1">
          <li>Generated values, seeds, or pasted content</li>
          <li>Personal information, IP addresses, or device fingerprints</li>
          <li>Cookies or tracking identifiers</li>
        </ul>
        <p className="text-sm leading-relaxed text-foreground mb-4">
          Umami is privacy-focused analytics software that complies with GDPR,
          CCPA, and PECR without requiring cookie consent banners.
        </p>

        <h2 className="font-mono text-lg text-foreground mt-8 mb-3">Do Not Track</h2>
        <p className="text-sm leading-relaxed text-foreground mb-4">
          If your browser sends a Do Not Track (DNT) header, Umami honors it
          and no analytics data is collected from your session.
        </p>

        <h2 className="font-mono text-lg text-foreground mt-8 mb-3">Local Installation</h2>
        <p className="text-sm leading-relaxed text-foreground mb-2">
          When you install Smedje locally via:
        </p>
        <p className="text-sm leading-relaxed text-foreground mb-4">
          <code className="font-mono bg-panel px-1">
            go install github.com/MydsiIversen/smedje/cmd/smedje@latest
          </code>
        </p>
        <p className="text-sm leading-relaxed text-foreground mb-4">
          No analytics of any kind are collected. The binary runs entirely
          on your machine with no network calls.
        </p>

        <h2 className="font-mono text-lg text-foreground mt-8 mb-3">Source Code</h2>
        <p className="text-sm leading-relaxed text-foreground mb-2">
          Smedje is open source under the AGPL-3.0 license. You can review
          the complete source code, including this privacy policy, at:{" "}
          <a
            href="https://github.com/MydsiIversen/smedje"
            target="_blank"
            rel="noopener noreferrer"
            className="font-mono bg-panel px-1 hover:text-forge transition-colors duration-150"
          >
            github.com/MydsiIversen/smedje
          </a>
        </p>
        <p className="text-sm leading-relaxed text-muted-foreground mt-6">
          Questions? Open an issue on GitHub.
        </p>
      </div>
    </div>
  )
}
