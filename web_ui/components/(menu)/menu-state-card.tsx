import { Card } from "@/components/ui/card";

type MenuStateCardProps = {
  message: string;
};

export function MenuStateCard({ message }: MenuStateCardProps) {
  return (
    <main className="min-h-screen bg-sl-f3f7f5 p-6">
      <div className="mx-auto max-w-[1240px]">
        <Card className="p-6 text-sm text-sl-607b74">{message}</Card>
      </div>
    </main>
  );
}
