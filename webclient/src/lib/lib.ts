const BASE_URL = "http://localhost:8080";

export async function connect(ip: string, port: string): Promise<boolean> {
  const res = await fetch(`${BASE_URL}/connect`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ ip, port }),
  });

  if (!res.ok) {
    throw new Error("Connection failed");
  }

  return true;
}

export async function ping(): Promise<string> {
  try {
    const response = await fetch(`${BASE_URL}/ping`);

    if (!response.ok) {
      throw new Error("Network response was not ok");
    }

    const data = await response.json();
    console.log(data);
    return data.reply;
  } catch (error) {
    console.error("There was a problem with the fetch operation:", error);
    throw error;
  }
}

export async function get(key: string): Promise<string> {
  const response = await fetch(`${BASE_URL}/getkey`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ key }),
  });

  if (!response.ok) {
    throw new Error("Network response was not ok");
  }

  const data = await response.json();
  console.log(data);
  return data.reply;
}

export async function set(key: string, value: string): Promise<string> {
  const response = await fetch(`${BASE_URL}/setkey`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ key, value }),
  });

  if (!response.ok) {
    throw new Error("Network response was not ok");
  }

  const data = await response.json();
  console.log(data);
  return data.reply;
}

export async function strln(key: string): Promise<string> {
  const response = await fetch(`${BASE_URL}/strln`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ key }),
  });

  if (!response.ok) {
    throw new Error("Network response was not ok");
  }

  const data = await response.json();
  console.log(data);
  return data.reply;
}

export async function del(key: string): Promise<string> {
  const response = await fetch(`${BASE_URL}/delkey`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ key }),
  });

  if (!response.ok) {
    throw new Error("Network response was not ok");
  }

  const data = await response.json();
  console.log(data);
  return data.reply;
}

export async function append(key: string, value: string): Promise<string> {
  const response = await fetch(`${BASE_URL}/append`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ key, value }),
  });

  if (!response.ok) {
    throw new Error("Network response was not ok");
  }

  const data = await response.json();
  console.log(data);
  return data.reply;
}

export type LogType = {
  term: number;
  command: {
    command: string,
    key: string,
    value: string
  };
};

export async function log(): Promise<LogType[]> {
  const response = await fetch(`${BASE_URL}/log`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!response.ok) {
    throw new Error("Network response was not ok");
  }

  const data = await response.json();
  return data.reply;
}
