import { createSignal } from "solid-js";
import { A } from "@solidjs/router";
import pb from "../lib/pb";

// Keyframes for the logo spin, scoped to this component via an inline
// <style> tag so style.css doesn't need to know about it.
const spinStyle = `
@keyframes logo-spin {
  to { transform: rotate(360deg); }
}
.logo-spin {
  animation: logo-spin 0.6s ease;
}
`;

export default function NavBar(props) {
  const [refreshing, setRefreshing] = createSignal(false);
  const [spinning, setSpinning] = createSignal(false);

  const handleLogout = () => pb.authStore.clear();

  // Clicking the logo resets the current page's state (if the page
  // passed one in via onLogoClick) and plays a one-turn spin. Clicking
  // from another route still navigates home via the surrounding <A>.
  const handleLogoClick = () => {
    props.onLogoClick?.();
    setSpinning(true);
  };

  return (
    <div class="mb-10 flex w-full flex-wrap items-center justify-between gap-y-3">
      <style>{spinStyle}</style>
      <A
        href="/"
        class="font-serif text-4xl flex items-center gap-2 transition-opacity hover:opacity-80"
        onClick={handleLogoClick}
      >
        <img
          src="/favicon.svg"
          alt=""
          class="h-12 w-12"
          classList={{ "logo-spin": spinning() }}
          onAnimationEnd={() => setSpinning(false)}
        />
        <h1>picmd</h1>
      </A>
      <nav class="flex flex-wrap items-center gap-3">
        <A href="/stats" class="btn">
          Stats
        </A>
        <A href="/settings" class="btn">
          Settings
        </A>
        <button type="button" class="btn" onClick={handleLogout}>
          Log out
        </button>
      </nav>
    </div>
  );
}
