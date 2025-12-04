import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { UserAvatar } from "@/components/ui/user-avatar";
import { getUsersOptions } from "@/lib/api/@tanstack/react-query.gen";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { Check, ChevronsUpDown } from "lucide-react";
import { useState } from "react";

interface UserFilterProps {
  value?: string;
  onChange: (value: string | undefined) => void;
}

export function UserFilter({ value, onChange }: UserFilterProps) {
  const [open, setOpen] = useState(false);

  const { data: response } = useQuery({
    ...getUsersOptions(),
  });

  const selectedUser = response?.users?.find((user) => user.id === value);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="min-w-[200px] max-w-full justify-between"
        >
          {selectedUser ? (
            <div className="flex items-center gap-2 min-w-0 flex-1">
              <UserAvatar userId={selectedUser.id} size="sm" />
            </div>
          ) : (
            <span className="truncate">Выберите пользователя...</span>
          )}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[var(--radix-popover-trigger-width)] max-w-[400px] p-0">
        <Command>
          <CommandInput placeholder="Поиск пользователя..." />
          <CommandEmpty>Пользователь не найден.</CommandEmpty>
          <CommandGroup className="max-h-[300px] overflow-y-auto">
            <CommandItem
              onSelect={() => {
                onChange(undefined);
                setOpen(false);
              }}
            >
              <Check
                className={cn(
                  "mr-2 h-4 w-4",
                  !value ? "opacity-100" : "opacity-0",
                )}
              />
              Все пользователи
            </CommandItem>
            {response?.users?.map((user) => (
              <CommandItem
                key={user.id}
                onSelect={() => {
                  onChange(user.id);
                  setOpen(false);
                }}
              >
                <div className="flex items-center gap-2">
                  <UserAvatar userId={user.id} size="sm" />
                  <span>{`${user.fullName}`}</span>
                </div>
                <Check
                  className={cn(
                    "ml-auto h-4 w-4",
                    value === user.id ? "opacity-100" : "opacity-0",
                  )}
                />
              </CommandItem>
            ))}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
