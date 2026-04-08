"use client";

import { useI18n } from "@/context/I18nContext";
import Link from "next/link";

export function DisclaimerFooter() {
  const { t } = useI18n();

  return (
    <footer className="border-t border-border-theme px-4 py-3 text-center">
      <p className="text-xs text-text-muted">
        {t("chat.disclaimer")}{" "}
        <Link href="/terms" className="text-indigo-400 hover:underline">
          {t("terms.title")}
        </Link>
      </p>
    </footer>
  );
}
