"use client";

import { useI18n } from "@/context/I18nContext";
import { ArrowLeft, Shield, FileText, Lock } from "lucide-react";
import Link from "next/link";

export default function TermsPage() {
  const { t, locale } = useI18n();

  return (
    <div className="max-w-3xl mx-auto py-16 px-6">
      <Link href="/dashboard" className="inline-flex items-center gap-1.5 text-sm text-indigo-400 hover:text-indigo-300 mb-8 transition">
        <ArrowLeft size={16} />
        {t("btn.back")}
      </Link>

      <h1 className="text-3xl font-bold text-foreground mb-8 flex items-center gap-3">
        <FileText size={28} className="text-indigo-400" />
        {t("terms.title")}
      </h1>

      <div className="space-y-8 text-sm text-text-muted leading-relaxed">
        {/* Disclaimer */}
        <section>
          <h2 className="text-lg font-semibold text-foreground mb-3 flex items-center gap-2">
            <Shield size={18} className="text-yellow-400" />
            {t("terms.disclaimer_title")}
          </h2>
          <p>{t("terms.disclaimer_text")}</p>
          <p className="mt-3">{t("terms.risk_warning")}</p>
        </section>

        {/* Terms of Use */}
        <section>
          <h2 className="text-lg font-semibold text-foreground mb-3">{t("terms.usage_title")}</h2>
          <ul className="list-disc list-inside space-y-2">
            <li>{t("terms.usage_1")}</li>
            <li>{t("terms.usage_2")}</li>
            <li>{t("terms.usage_3")}</li>
          </ul>
        </section>

        {/* Privacy Policy */}
        <section>
          <h2 className="text-lg font-semibold text-foreground mb-3 flex items-center gap-2">
            <Lock size={18} className="text-green-400" />
            {t("terms.privacy_title")}
          </h2>
          <p>{t("terms.privacy_text")}</p>
          <ul className="list-disc list-inside space-y-2 mt-3">
            <li>{t("terms.privacy_1")}</li>
            <li>{t("terms.privacy_2")}</li>
            <li>{t("terms.privacy_3")}</li>
          </ul>
        </section>

        {/* Contact */}
        <section>
          <h2 className="text-lg font-semibold text-foreground mb-3">{t("terms.contact_title")}</h2>
          <p>{t("terms.contact_text")}</p>
        </section>
      </div>
    </div>
  );
}
