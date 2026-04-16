import React from "react";

const ACCESS_LINKS = {
  worker: "https://github.com/Shravanthi20/InDel#worker-app", // Android only
  platform: "https://indel-platform-dashboard.onrender.com",
  insurer: "https://indel-urn9.onrender.com",
};

export default function HomePage() {
  return (
    <main className="bg-white text-gray-900 font-sans">
      {/* Hero Section */}
      <section className="w-full min-h-[80vh] flex flex-col justify-center items-center bg-gradient-to-b from-indigo-50 to-white pb-16 pt-24">
        <h1 className="text-4xl md:text-6xl font-bold text-indigo-700 text-center max-w-3xl leading-tight">
          AI-Powered Income Protection for Gig Workers
        </h1>
        <p className="mt-6 text-lg md:text-2xl text-gray-700 max-w-xl text-center">
          When the rain falls, the orders stop. InDel protects delivery workers with automated, parametric insurance—no paperwork, no hassle.
        </p>
        <div className="mt-8 flex gap-4">
          <a href="#demo" className="px-6 py-3 rounded-lg bg-indigo-600 text-white font-semibold shadow hover:bg-indigo-700 transition">Watch Demo</a>
          <a href="#ecosystem" className="px-6 py-3 rounded-lg bg-white border border-indigo-600 text-indigo-700 font-semibold shadow hover:bg-indigo-50 transition">Explore Product</a>
        </div>
      </section>

      {/* Problem Section */}
      <section className="max-w-4xl mx-auto py-16 px-4" id="problem">
        <h2 className="text-2xl md:text-3xl font-bold mb-4 text-indigo-700">The Problem</h2>
        <p className="text-lg text-gray-700 mb-4">
          India has 15+ million gig delivery workers. Their income is fragile—one flood, a pollution spike, or a citywide curfew, and earnings collapse overnight. There’s no fallback. No insurance product covers these real-world disruptions.
        </p>
        <ul className="list-disc ml-6 text-gray-700 space-y-2">
          <li>Weather events (rain, flood, heatwaves) halt deliveries</li>
          <li>Pollution and AQI spikes trigger bans</li>
          <li>Curfews and civic disruptions stop work instantly</li>
          <li>20–30% of monthly income lost during such events</li>
        </ul>
      </section>

      {/* Solution Section */}
      <section className="max-w-4xl mx-auto py-16 px-4" id="solution">
        <h2 className="text-2xl md:text-3xl font-bold mb-4 text-indigo-700">Our Solution: InDel</h2>
        <p className="text-lg text-gray-700 mb-4">
          InDel is an AI-powered parametric insurance platform. We offer weekly pricing, instant onboarding, and fully automated payouts—no claims to file, no paperwork, no friction.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-8">
          <div className="bg-indigo-50 rounded-xl p-6 shadow flex flex-col">
            <span className="text-3xl font-bold text-indigo-600">₹10–15/week</span>
            <span className="mt-2 text-gray-700">Typical premium for full coverage</span>
          </div>
          <div className="bg-indigo-50 rounded-xl p-6 shadow flex flex-col">
            <span className="text-3xl font-bold text-indigo-600">Zero paperwork</span>
            <span className="mt-2 text-gray-700">Automated, instant payouts when disruption is detected</span>
          </div>
        </div>
      </section>

      {/* How It Works Section */}
      <section className="max-w-5xl mx-auto py-16 px-4" id="how-it-works">
        <h2 className="text-2xl md:text-3xl font-bold mb-8 text-indigo-700">How It Works</h2>
        <div className="grid md:grid-cols-4 gap-6">
          {[
            {
              title: "1. Register & Onboard",
              desc: "Worker signs up in the app and selects a plan."
            },
            {
              title: "2. AI Risk Scoring",
              desc: "ML model calculates fair premium based on real data."
            },
            {
              title: "3. Disruption Detection",
              desc: "Weather, AQI, or civic triggers are monitored in real time."
            },
            {
              title: "4. Automated Payout",
              desc: "If disruption hits, payout is credited instantly—no claim needed."
            }
          ].map((step, i) => (
            <div key={i} className="bg-white rounded-xl shadow p-6 flex flex-col items-center text-center border-t-4 border-indigo-500">
              <span className="text-3xl font-bold text-indigo-600 mb-2">{step.title}</span>
              <span className="text-gray-700">{step.desc}</span>
            </div>
          ))}
        </div>
      </section>

      {/* Technology Layer Section */}
      <section className="max-w-5xl mx-auto py-16 px-4" id="tech">
        <h2 className="text-2xl md:text-3xl font-bold mb-8 text-indigo-700">Technology Layer</h2>
        <div className="grid md:grid-cols-4 gap-6">
          {[
            { title: "AI Risk Scoring", desc: "Personalized pricing using ML models" },
            { title: "Parametric Triggers", desc: "Weather, AQI, and event APIs" },
            { title: "Fraud Detection", desc: "Multi-layer ML fraud stack" },
            { title: "Claims Automation", desc: "No manual claims—payouts are instant" },
          ].map((tech, i) => (
            <div key={i} className="bg-indigo-50 rounded-xl shadow p-6 flex flex-col items-center text-center">
              <span className="text-xl font-bold text-indigo-700 mb-2">{tech.title}</span>
              <span className="text-gray-700">{tech.desc}</span>
            </div>
          ))}
        </div>
      </section>

      {/* Product Ecosystem Section */}
      <section className="max-w-5xl mx-auto py-16 px-4" id="ecosystem">
        <h2 className="text-2xl md:text-3xl font-bold mb-8 text-indigo-700">Product Ecosystem</h2>
        <div className="grid md:grid-cols-3 gap-6">
          {[
            {
              title: "Worker App",
              desc: "Android app for gig workers",
              link: ACCESS_LINKS.worker,
              btn: "View on GitHub"
            },
            {
              title: "Platform Dashboard",
              desc: "Operations dashboard for platforms",
              link: ACCESS_LINKS.platform,
              btn: "Open Dashboard"
            },
            {
              title: "Insurer Dashboard",
              desc: "Claims and risk dashboard for insurers",
              link: ACCESS_LINKS.insurer,
              btn: "Open Dashboard"
            }
          ].map((item, i) => (
            <div key={i} className="bg-white rounded-xl shadow p-8 flex flex-col items-center text-center border-t-4 border-indigo-500">
              <span className="text-xl font-bold text-indigo-700 mb-2">{item.title}</span>
              <span className="text-gray-700 mb-4">{item.desc}</span>
              <a href={item.link} target="_blank" rel="noopener noreferrer" className="mt-auto px-5 py-2 rounded-lg bg-indigo-600 text-white font-semibold shadow hover:bg-indigo-700 transition">{item.btn}</a>
            </div>
          ))}
        </div>
      </section>

      {/* Dashboard Preview Section */}
      <section className="max-w-5xl mx-auto py-16 px-4" id="dashboard-preview">
        <h2 className="text-2xl md:text-3xl font-bold mb-8 text-indigo-700">Dashboard Preview</h2>
        <div className="grid md:grid-cols-4 gap-6">
          {[
            { label: "Earnings Protected", value: "₹44,000" },
            { label: "Weekly Premium", value: "₹10–15" },
            { label: "Claims Triggered", value: "1,200" },
            { label: "Avg. Risk Score", value: "0.42" },
          ].map((stat, i) => (
            <div key={i} className="bg-indigo-50 rounded-xl shadow p-8 flex flex-col items-center text-center">
              <span className="text-3xl font-bold text-indigo-700 mb-2">{stat.value}</span>
              <span className="text-gray-700">{stat.label}</span>
            </div>
          ))}
        </div>
      </section>

      {/* Why We Built This Section */}
      <section className="max-w-4xl mx-auto py-16 px-4" id="why">
        <h2 className="text-2xl md:text-3xl font-bold mb-4 text-indigo-700">Why We Built This</h2>
        <p className="text-lg text-gray-700 mb-4">
          Our team has seen firsthand how gig workers struggle with unpredictable income. Traditional insurance doesn’t fit their needs, and platforms rarely step in. We built InDel to give workers a safety net that’s fair, affordable, and truly automated—so they can focus on work, not worry.
        </p>
      </section>

      {/* Vision / Future Scope Section */}
      <section className="max-w-4xl mx-auto py-16 px-4" id="vision">
        <h2 className="text-2xl md:text-3xl font-bold mb-4 text-indigo-700">Vision & Future Scope</h2>
        <p className="text-lg text-gray-700 mb-4">
          InDel is built to scale—across cities, platforms, and worker types. Our vision is to become the default income protection layer for the gig economy, expanding to new geographies and risk types.
        </p>
      </section>

      {/* Demo Section */}
      <section className="max-w-4xl mx-auto py-16 px-4" id="demo">
        <h2 className="text-2xl md:text-3xl font-bold mb-8 text-indigo-700">Product Demo</h2>
        <div className="aspect-w-16 aspect-h-9 w-full rounded-xl overflow-hidden shadow-lg mb-6">
          <iframe
            src="https://www.youtube.com/embed/R1_1X-f7-MM"
            title="InDel Demo"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowFullScreen
            className="w-full h-96 border-none"
          />
        </div>
        <p className="text-gray-700 text-center max-w-2xl mx-auto">
          See how InDel works in action: onboarding, disruption detection, and automated payouts—all in under 2 minutes.
        </p>
      </section>
    </main>
  );
}
