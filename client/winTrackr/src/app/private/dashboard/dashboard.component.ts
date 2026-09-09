import { Component, OnInit } from '@angular/core';
import { AuthService } from '../../core/services/auth.service';
import { ScratchOffEntry, ScratchOffEntryService, ScratchoffFullData } from '../../core/services/scratch-off-entry.service';
import { Router } from '@angular/router';
import { MatTableDataSource } from '@angular/material/table';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { AlertService } from 'app/core/services/alert.service';

export interface Entry {
  entryId: number | undefined;
  date: string | undefined;
  id: string | undefined;
  name: string | undefined;
  winAmount: number | undefined;
}

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class DashboardComponent implements OnInit {
  message: string = "Loading...";
  displayedColumns: string[] = ['date', 'id', 'name', 'winAmount', 'actions'];
  entries: MatTableDataSource<Entry>;
  entryForm: FormGroup;
  editingEntryId: number | undefined;

  constructor(
    private authService: AuthService,
    private scratchOffEntryService: ScratchOffEntryService,
    private router: Router,
    private formBuilder: FormBuilder, 
    private alertService: AlertService,
  ) {
    this.entries = new MatTableDataSource<Entry>([]);
    this.entryForm = this.formBuilder.group({
      name: ['', Validators.required],
      id: ['', Validators.required],
      winAmount: ['', Validators.required],
      date: ['', Validators.required]
    });
  };

  ngOnInit(): void {
    this.authService.userLanding()
      .subscribe({
        next: (response: any) => {
          this.message = response;
          // Fetch initial entries here
          this.refreshEntries();
        },
        error: (error: any) => {
          this.message = "Error";
          this.router.navigate(['/login']);
          this.alertService.showAlert('Unauthorized: Please login');
        }
      });
  }

  refreshEntries(): void {
    this.scratchOffEntryService.getEntries()
      .subscribe((res)=>{
        var resEntries = res as ScratchoffFullData[]
        if (resEntries == null) {
          return
        }
        const entries: Entry[] = [];
        for (let i = 0; i < resEntries.length; i++) {
          var entry = resEntries.at(i);
          entries.push({ entryId: entry?.Id, date: entry?.Date, id: entry?.GameId, name: entry?.Name, winAmount: entry?.Prize });
        }
        this.entries.data = entries;
      });
  }

  onSubmit(): void {
    const newEntry: Entry = {
      entryId: this.editingEntryId,
      date: this.entryForm.get('date')?.value,
      id: this.entryForm.get('id')?.value,
      name: this.entryForm.get('name')?.value,
      winAmount: this.entryForm.get('winAmount')?.value,
    };

    if (this.editingEntryId !== undefined) {
      this.scratchOffEntryService.updateEntry(this.editingEntryId, newEntry)
        .subscribe({
          next: () => {
            this.cancelEdit();
            this.refreshEntries();
          },
          error: () => {
            this.alertService.showAlert('Failed to update entry. Please try again later.');
          }
        });
      return;
    }

    // Send the newEntry data to the backend to save it in the user's data
    this.scratchOffEntryService.addEntry(newEntry);

    // After sending the data to the backend, you can clear the form and refresh the entries
    this.clearForm();
    this.refreshEntries();
  }

  clearForm(): void {
  this.entryForm.reset();
  }

  // Populate the form with an existing entry's data and switch the form to update mode
  editEntry(entry: Entry): void {
    this.editingEntryId = entry.entryId;
    this.entryForm.setValue({
      name: entry.name,
      id: entry.id,
      winAmount: entry.winAmount,
      date: entry.date,
    });
  }

  // Exit update mode without saving changes
  cancelEdit(): void {
    this.editingEntryId = undefined;
    this.clearForm();
  }

  deleteEntry(entry: Entry): void {
    if (entry.entryId === undefined) {
      return;
    }
    if (!confirm(`Delete the "${entry.name}" entry?`)) {
      return;
    }

    this.scratchOffEntryService.deleteEntry(entry.entryId)
      .subscribe({
        next: () => {
          if (this.editingEntryId === entry.entryId) {
            this.cancelEdit();
          }
          this.refreshEntries();
        },
        error: () => {
          this.alertService.showAlert('Failed to delete entry. Please try again later.');
        }
      });
  }

  // Event handler for generating tax report
  generateTaxReport(): void {
    console.error('Generate Tax Report clicked.');
    this.alertService.showAlert('A server error occurred. Please try again later.');
  }

  // Event handler for logging out
  logout(): void {
    console.error('Logout clicked.');
    this.authService.logout();
    this.router.navigate(['/']);
  }
  
  }