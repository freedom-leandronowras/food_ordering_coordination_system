"use client";

import { useMenuContext } from "@/components/(menu)/menu-context";
import type { MenuSectionsData } from "@/components/(menu)/menu-section-types";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { formatMoney } from "@/lib/menu-data";

type MainSectionProps = {
  featured: MenuSectionsData["featured"];
  items: MenuSectionsData["items"];
};

export function MainSection({ featured, items }: MainSectionProps) {
  const { isBootstrapping, selectedVendor, vendors, selectedVendorId, selectVendor, addToCart } =
    useMenuContext();
  const featuredDescription = selectedVendor
    ? featured.vendorDescriptions[selectedVendor.service_name] ?? `Popular picks from ${selectedVendor.service_name}.`
    : featured.description;

  return (
    <div className="space-y-5">
      <Card className="rounded-[28px] border-[#d6e7e1] bg-gradient-to-r from-[#123830] via-[#1b5b52] to-[#286e63] p-6 text-white">
        <h2 className="text-3xl font-semibold leading-tight md:text-4xl">
          {selectedVendor?.service_name ?? "Loading vendor..."}
        </h2>
        <p className="mt-2 max-w-xl text-sm text-[#d4ebe5]">{featuredDescription}</p>
      </Card>

      <Card className="p-4">
        <div className="flex flex-wrap gap-2">
          {vendors.map((vendor) => (
            <Button
              key={vendor.service_id}
              type="button"
              variant={selectedVendorId === vendor.service_id ? "default" : "outline"}
              size="sm"
              onClick={() => selectVendor(vendor.service_id)}
            >
              {vendor.service_name}
            </Button>
          ))}
        </div>
      </Card>

      <div className="space-y-3">
        <h3 className="text-2xl font-semibold md:text-3xl">{items.title}</h3>
        {isBootstrapping ? (
          <Card className="p-5 text-sm text-[#607b74]">{items.loadingText}</Card>
        ) : (
          <div className="space-y-3">
            {selectedVendor?.items.map((item) => (
              <Card key={item.id} className="p-4">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <p className="text-lg font-semibold md:text-xl">{item.name}</p>
                    <p className="mt-1 text-sm text-[#5a746d]">{item.description}</p>
                  </div>
                  <div className="text-right">
                    <p className="text-xl font-semibold text-[#1c5f54]">{formatMoney(item.price)}</p>
                    <Button
                      type="button"
                      onClick={() => addToCart(selectedVendor?.service_name ?? "Vendor", item)}
                      disabled={!item.available}
                      className="mt-3"
                      size="sm"
                    >
                      {item.available ? "Add to card" : "Unavailable"}
                    </Button>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
